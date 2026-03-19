job "{{ .JobName }}" {
  datacenters = ["dc1"]
  type = "service"

  group "zone" {
    restart {
        attempts = 10
        interval = "5m"
        delay    = "10s"
        mode     = "delay"
      }

    network {
      mode = "bridge"

      port "game" {
        to = 7777
      }
    }

    task "zone" {
      driver = "docker"

      config {
        image = "{{ .ImageName }}"
        ports = ["game"]
        args  = ["--world-id={{ .WorldID }}", "--zone-id={{ .ZoneID }}"]
      }

      service {
        name = "zone-server"
        port = "game"
        tags = [
          "world-{{ .WorldID }}",
          "zone-{{ .ZoneID }}"
        ]

        check {
          type     = "tcp"
          interval = "10s"
          timeout  = "3s"
        }
      }
    }
  }
}
