plans
- id
- name
- cpu
- ram_mb
- disk_gb
- price_month
- is_active
- created_at


vds
- id
- user_id          -- ID из auth-service
- plan_id
- node_id          -- на каком proxmox-ноде
- proxmox_vm_id
- status           -- creating | running | stopped | error | deleting
- ipv4
- ipv6
- created_at
- expires_at


nodes
- id
- name
- api_url
- max_cpu
- max_ram
- max_disk
- is_active


tasks
- id
- vds_id
- type            -- create | delete | start | stop
- status          -- pending | running | done | error
- error
- created_at