// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/kdomanski/iso9660"
	"os"
	"strings"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &UnattendedISOResource{}
var _ resource.ResourceWithImportState = &UnattendedISOResource{}

func NewUnattendedISOResource() resource.Resource {
	return &UnattendedISOResource{}
}

// UnattendedISOResource defines the resource implementation.
type UnattendedISOResource struct {
}

// UnattendedISOResourceModel describes the resource data model.
type UnattendedISOResourceModel struct {
	Id           types.String `tfsdk:"id"`
	FileName     types.String `tfsdk:"file_name"`
	PathOverride types.String `tfsdk:"path_override"`
	XMLContent   types.String `tfsdk:"xml_content"`
	ResultPath   types.String `tfsdk:"result_path"`
}

func (r *UnattendedISOResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "unattend_iso_file"
}

func (r *UnattendedISOResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Unattend ISO Resource.",

		Attributes: map[string]schema.Attribute{
			"path_override": schema.StringAttribute{
				MarkdownDescription: "Path to write the local ISO file, defaults to OS temp",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString("tmp"),
			},
			"file_name": schema.StringAttribute{
				MarkdownDescription: "Name for the created ISO file",
				Optional:            false,
				Required:            true,
			},
			"xml_content": schema.StringAttribute{
				MarkdownDescription: "XML content for the unattend.xml file.",
				Optional:            false,
				Required:            true,
			},
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "ISO identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"result_path": schema.StringAttribute{
				Computed:            true,
				Required:            false,
				Optional:            false,
				MarkdownDescription: "Resultant File Path",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *UnattendedISOResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

}

func (r *UnattendedISOResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data UnattendedISOResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
	//     return
	// }

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	data.Id = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	isoWriter, err := iso9660.NewWriter()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to start ISO Writer, got error: %s", err))
		return
	}
	defer func(isoWriter *iso9660.ImageWriter) {
		err := isoWriter.Cleanup()
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error in ISO Writer, got error: %s", err))
			return
		}
	}(isoWriter)

	if data.XMLContent.String() != "" {
		err = isoWriter.AddFile(strings.NewReader(data.XMLContent.String()), "unattend.xml")
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error adding file to ISO, got error: %s", err))
			return
		}
	}

	var b bytes.Buffer
	err = isoWriter.WriteTo(&b, "unattend")
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error writing to ISO, got error: %s", err))
		return
	}

	// Calculate the ISO sha256 sum
	//sum := fmt.Sprintf("%x", sha256.Sum256(b.Bytes()))

	if data.PathOverride.String() != "tmp" {
		file, err := os.CreateTemp("/tmp", data.FileName.String())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error creating tmp file, got error: %s", err))
			return
		}
		_, err = file.Write(b.Bytes())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error writing to tmp file, got error: %s", err))
			return
		}
		data.ResultPath = types.StringValue(file.Name())
	} else {
		file, err := os.Create(data.PathOverride.String() + data.FileName.String())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error creating file, got error: %s", err))
			return
		}
		_, err = file.Write(b.Bytes())
		if err != nil {
			resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Error writing file, got error: %s", err))
			return
		}
		data.ResultPath = types.StringValue(file.Name())
	}

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UnattendedISOResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data UnattendedISOResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UnattendedISOResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data UnattendedISOResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update example, got error: %s", err))
	//     return
	// }

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *UnattendedISOResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data UnattendedISOResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// If applicable, this is a great opportunity to initialize any necessary
	// provider client data and make a call using it.
	// httpResp, err := r.client.Do(httpReq)
	// if err != nil {
	//     resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete example, got error: %s", err))
	//     return
	// }
}

func (r *UnattendedISOResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
