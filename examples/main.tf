terraform {
  required_providers {
    unattend = {
      source = "Br4v3St4rr/unattend-iso"
    }
  }
}

resource "unattend_iso_file" "test" {

}