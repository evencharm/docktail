# Docktail

Container based on `promtail` to spit out a static JSON file to use as a file based discovery.

```yaml
scrape_configs:
- job_name: containers
  entry_parser: docker
  file_sd_configs:
  - files:
    - /etc/promtail/promtail-targets.json
  relabel_configs:
  - source_labels: [__address__]
    target_label: container_id
  - source_labels: [container_id]
    target_label: __path__
    replacement: /var/lib/docker/containers/$1*/*.log
```
