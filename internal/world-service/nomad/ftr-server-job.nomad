job "{{ .JobName }}" {
  datacenters = ["dc1"]
  type = "service"

  group "zone" {
    restart {
      attempts = 0
      interval = "24h"
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
    }
  }
}
