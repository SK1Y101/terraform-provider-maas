resource "maas_boot_source_selection" "test" {
  boot_source = maas_boot_source.example.id

  os      = "ubuntu"
  release = "jammy"
}
