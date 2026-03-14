job "{{ .JobName }}" {
  datacenters = ["dc1"]
  type = "service"

  group "zone" {

    network {
      port "game" {
        to       = 7777
        protocol = "udp"
      }
    }

    task "zone" {
      driver = "docker"

      config {
        image = "{{ .ImageName }}"
        ports = ["game"]
        args  = ["--world-id={{ .WorldID }}", "--zone-id={{ .ZoneID }}"]
      }

      resources {
        cpu    = 200
        memory = 256
      }
    }
  }
}
