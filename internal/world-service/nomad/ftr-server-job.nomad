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

      port "health" {
        to = 7778
      }
    }

    task "zone" {
      driver = "docker"

      config {
        image = "{{ .ImageName }}"
        ports = ["game", "health"]
        args  = ["--world-id={{ .WorldID }}", "--zone-id={{ .ZoneID }}"]
      }

      meta {
        deployed_at = "{{ .DeployedAt }}"
      }

      service {
        name = "zone-server"
        port = "game"
        address_mode = "host"

        tags = [
          "world-{{ .WorldID }}",
          "zone-{{ .ZoneID }}"
        ]

        meta {
          public_ip = "${attr.unique.platform.aws.public-ipv4}"
        }

        check {
          type     = "tcp"
          port         = "health"
          address_mode = "host"
          interval = "10s"
          timeout  = "3s"
        }
      }
    }
  }
}
