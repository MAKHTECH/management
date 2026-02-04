-- ============================================================================
-- VDS Management Service - Database Schema
-- ============================================================================

-- Drop tables if they exist (in correct order due to foreign keys)
DROP TABLE IF EXISTS tasks CASCADE;
DROP TABLE IF EXISTS vds CASCADE;
DROP TABLE IF EXISTS plans CASCADE;
DROP TABLE IF EXISTS nodes CASCADE;

-- ============================================================================
-- PLANS TABLE
-- ============================================================================
CREATE TABLE plans (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(100) NOT NULL UNIQUE,
                       cpu INTEGER NOT NULL CHECK (cpu > 0),
                       ram_mb INTEGER NOT NULL CHECK (ram_mb > 0),
                       disk_gb INTEGER NOT NULL CHECK (disk_gb > 0),
                       price_month DECIMAL(10, 2) NOT NULL CHECK (price_month >= 0),
                       is_active BOOLEAN NOT NULL DEFAULT true,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_plans_is_active ON plans(is_active);
CREATE INDEX idx_plans_price ON plans(price_month) WHERE is_active = true;

COMMENT ON TABLE plans IS 'VDS tariff plans with resource allocations';
COMMENT ON COLUMN plans.ram_mb IS 'RAM in megabytes';
COMMENT ON COLUMN plans.disk_gb IS 'Disk space in gigabytes';
COMMENT ON COLUMN plans.price_month IS 'Monthly price in RUB (КОПЕЙКИ)';

-- ============================================================================
-- NODES TABLE
-- ============================================================================
CREATE TABLE nodes (
                       id SERIAL PRIMARY KEY,
                       name VARCHAR(100) NOT NULL UNIQUE,
                       api_url VARCHAR(255) NOT NULL,
                       max_cpu INTEGER NOT NULL CHECK (max_cpu > 0),
                       max_ram INTEGER NOT NULL CHECK (max_ram > 0),
                       max_disk INTEGER NOT NULL CHECK (max_disk > 0),
                       is_active BOOLEAN NOT NULL DEFAULT true,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_nodes_is_active ON nodes(is_active);

COMMENT ON TABLE nodes IS 'Proxmox nodes (physical servers)';
COMMENT ON COLUMN nodes.api_url IS 'Proxmox API endpoint URL';
COMMENT ON COLUMN nodes.max_ram IS 'Maximum RAM in MB';
COMMENT ON COLUMN nodes.max_disk IS 'Maximum disk space in GB';

-- ============================================================================
-- VDS TABLE
-- ============================================================================
CREATE TABLE vds (
                     id SERIAL PRIMARY KEY,
                     user_id INTEGER NOT NULL,
                     plan_id INTEGER NOT NULL REFERENCES plans(id) ON DELETE RESTRICT,
                     node_id INTEGER NOT NULL REFERENCES nodes(id) ON DELETE RESTRICT,
                     proxmox_vm_id INTEGER NOT NULL,
                     status VARCHAR(20) NOT NULL CHECK (
                         status IN ('creating', 'running', 'stopped', 'error', 'deleting')
                         ),
                     ipv4 INET,
                     ipv6 INET,
                     created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                     expires_at TIMESTAMP WITH TIME ZONE NOT NULL,

                     CONSTRAINT unique_proxmox_vm UNIQUE (node_id, proxmox_vm_id)
);

CREATE INDEX idx_vds_user_id ON vds(user_id);
CREATE INDEX idx_vds_status ON vds(status);
CREATE INDEX idx_vds_plan_id ON vds(plan_id);
CREATE INDEX idx_vds_node_id ON vds(node_id);
CREATE INDEX idx_vds_expires_at ON vds(expires_at);

COMMENT ON TABLE vds IS 'Virtual Dedicated Servers (VDS instances)';
COMMENT ON COLUMN vds.user_id IS 'User ID from external auth-service';
COMMENT ON COLUMN vds.proxmox_vm_id IS 'VM ID in Proxmox';
COMMENT ON COLUMN vds.status IS 'Current VDS state';
COMMENT ON COLUMN vds.expires_at IS 'Subscription expiration date';

-- ============================================================================
-- TASKS TABLE
-- ============================================================================
CREATE TABLE tasks (
                       id SERIAL PRIMARY KEY,
                       vds_id INTEGER NOT NULL REFERENCES vds(id) ON DELETE CASCADE,
                       type VARCHAR(20) NOT NULL CHECK (
                           type IN ('create', 'delete', 'start', 'stop', 'restart')
                           ),
                       status VARCHAR(20) NOT NULL CHECK (
                           status IN ('pending', 'running', 'done', 'error')
                           ),
                       error TEXT,
                       created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       started_at TIMESTAMP WITH TIME ZONE,
                       completed_at TIMESTAMP WITH TIME ZONE
);

CREATE INDEX idx_tasks_vds_id ON tasks(vds_id);
CREATE INDEX idx_tasks_status ON tasks(status);
CREATE INDEX idx_tasks_type ON tasks(type);
CREATE INDEX idx_tasks_created_at ON tasks(created_at DESC);

COMMENT ON TABLE tasks IS 'Background tasks for VDS operations';
COMMENT ON COLUMN tasks.type IS 'Type of operation to perform';
COMMENT ON COLUMN tasks.status IS 'Current task status';
COMMENT ON COLUMN tasks.error IS 'Error message if task failed';

-- ============================================================================
-- INSERT TEST DATA
-- ============================================================================

-- Insert Plans
INSERT INTO plans (name, cpu, ram_mb, disk_gb, price_month, is_active) VALUES
                       ('Starter', 1, 1024, 20, 5.00, true),
                       ('Basic', 2, 2048, 40, 10.00, true),
                       ('Standard', 4, 4096, 80, 20.00, true),
                       ('Advanced', 6, 8192, 160, 40.00, true),
                       ('Pro', 8, 16384, 320, 80.00, true),
                       ('Enterprise', 16, 32768, 640, 150.00, true),
                       ('Legacy Small', 1, 512, 10, 3.00, false);

-- Insert Nodes
INSERT INTO nodes (name, api_url, max_cpu, max_ram, max_disk, is_active) VALUES
                        ('node-eu-01', 'https://pve-eu-01.example.com:8006/api2/json', 64, 524288, 10000, true),
                        ('node-eu-02', 'https://pve-eu-02.example.com:8006/api2/json', 64, 524288, 10000, true),
                        ('node-us-01', 'https://pve-us-01.example.com:8006/api2/json', 128, 1048576, 20000, true),
                        ('node-asia-01', 'https://pve-asia-01.example.com:8006/api2/json', 64, 524288, 10000, true),
                         ('node-dev-01', 'https://pve-dev-01.example.com:8006/api2/json', 32, 131072, 5000, false);

-- Insert VDS instances
INSERT INTO vds (user_id, plan_id, node_id, proxmox_vm_id, status, ipv4, ipv6, created_at, expires_at) VALUES
                       (1001, 3, 1, 100, 'running', '185.22.45.101', '2a01:4f8:c17:1b4::1', '2024-01-15 10:30:00+00', '2025-01-15 10:30:00+00'),
                       (1001, 2, 1, 101, 'stopped', '185.22.45.102', '2a01:4f8:c17:1b4::2', '2024-02-20 14:15:00+00', '2025-02-20 14:15:00+00'),
                       (1002, 5, 2, 200, 'running', '195.88.73.50', '2a01:4f9:2b:3c1::1', '2024-03-10 09:00:00+00', '2025-03-10 09:00:00+00'),
                       (1003, 1, 1, 102, 'running', '185.22.45.103', '2a01:4f8:c17:1b4::3', '2024-06-05 16:45:00+00', '2025-06-05 16:45:00+00'),
                       (1004, 4, 3, 300, 'running', '192.168.100.10', '2001:db8:1234:5678::1', '2024-07-22 11:20:00+00', '2025-07-22 11:20:00+00'),
                       (1005, 6, 3, 301, 'running', '192.168.100.11', '2001:db8:1234:5678::2', '2024-08-30 13:00:00+00', '2025-08-30 13:00:00+00'),
                       (1002, 2, 2, 201, 'creating', NULL, NULL, '2025-02-03 10:00:00+00', '2026-02-03 10:00:00+00'),
                       (1006, 3, 4, 400, 'running', '103.28.54.120', '2404:6800:4003:c00::1', '2024-12-01 08:30:00+00', '2025-12-01 08:30:00+00'),
                       (1007, 1, 1, 103, 'error', '185.22.45.104', '2a01:4f8:c17:1b4::4', '2025-01-28 15:00:00+00', '2026-01-28 15:00:00+00');

-- Insert Tasks
INSERT INTO tasks (vds_id, type, status, error, created_at, started_at, completed_at) VALUES
                        (1, 'create', 'done', NULL, '2024-01-15 10:30:00+00', '2024-01-15 10:30:05+00', '2024-01-15 10:32:30+00'),
                        (1, 'start', 'done', NULL, '2024-01-15 10:32:35+00', '2024-01-15 10:32:36+00', '2024-01-15 10:33:00+00'),
                        (2, 'create', 'done', NULL, '2024-02-20 14:15:00+00', '2024-02-20 14:15:05+00', '2024-02-20 14:17:20+00'),
                        (2, 'stop', 'done', NULL, '2025-01-20 09:00:00+00', '2025-01-20 09:00:05+00', '2025-01-20 09:00:25+00'),
                        (3, 'create', 'done', NULL, '2024-03-10 09:00:00+00', '2024-03-10 09:00:05+00', '2024-03-10 09:03:15+00'),
                        (4, 'create', 'done', NULL, '2024-06-05 16:45:00+00', '2024-06-05 16:45:05+00', '2024-06-05 16:47:00+00'),
                        (5, 'create', 'done', NULL, '2024-07-22 11:20:00+00', '2024-07-22 11:20:05+00', '2024-07-22 11:22:45+00'),
                        (6, 'create', 'done', NULL, '2024-08-30 13:00:00+00', '2024-08-30 13:00:05+00', '2024-08-30 13:03:20+00'),
                        (7, 'create', 'running', NULL, '2025-02-03 10:00:00+00', '2025-02-03 10:00:05+00', NULL),
                        (8, 'create', 'done', NULL, '2024-12-01 08:30:00+00', '2024-12-01 08:30:05+00', '2024-12-01 08:32:15+00'),
                        (9, 'create', 'error', 'Failed to allocate IP address: no free IPs in pool', '2025-01-28 15:00:00+00', '2025-01-28 15:00:05+00', '2025-01-28 15:00:30+00'),
                        (3, 'restart', 'done', NULL, '2025-02-01 14:30:00+00', '2025-02-01 14:30:05+00', '2025-02-01 14:30:45+00'),
                        (1, 'stop', 'pending', NULL, '2025-02-03 11:00:00+00', NULL, NULL);

-- ============================================================================
-- VIEWS (Optional but useful for seniors)
-- ============================================================================

-- Active VDS with full plan details
CREATE OR REPLACE VIEW vds_active AS
SELECT
    v.id,
    v.user_id,
    v.proxmox_vm_id,
    v.status,
    v.ipv4,
    v.ipv6,
    v.created_at,
    v.expires_at,
    p.name as plan_name,
    p.cpu,
    p.ram_mb,
    p.disk_gb,
    p.price_month,
    n.name as node_name,
    n.api_url as node_api
FROM vds v
         JOIN plans p ON v.plan_id = p.id
         JOIN nodes n ON v.node_id = n.id
WHERE v.status IN ('running', 'stopped');

COMMENT ON VIEW vds_active IS 'Active VDS instances with plan and node details';

-- Node utilization statistics
CREATE OR REPLACE VIEW node_utilization AS
SELECT
    n.id,
    n.name,
    n.max_cpu,
    n.max_ram,
    n.max_disk,
    COUNT(v.id) as vds_count,
    COALESCE(SUM(p.cpu), 0) as used_cpu,
    COALESCE(SUM(p.ram_mb), 0) as used_ram,
    COALESCE(SUM(p.disk_gb), 0) as used_disk,
    ROUND(100.0 * COALESCE(SUM(p.cpu), 0) / n.max_cpu, 2) as cpu_usage_pct,
    ROUND(100.0 * COALESCE(SUM(p.ram_mb), 0) / n.max_ram, 2) as ram_usage_pct,
    ROUND(100.0 * COALESCE(SUM(p.disk_gb), 0) / n.max_disk, 2) as disk_usage_pct
FROM nodes n
         LEFT JOIN vds v ON n.id = v.node_id AND v.status IN ('running', 'stopped', 'creating')
         LEFT JOIN plans p ON v.plan_id = p.id
WHERE n.is_active = true
GROUP BY n.id, n.name, n.max_cpu, n.max_ram, n.max_disk;

COMMENT ON VIEW node_utilization IS 'Resource utilization statistics per node';

-- ============================================================================
-- USEFUL FUNCTIONS
-- ============================================================================

-- Function to get pending tasks count for a VDS
CREATE OR REPLACE FUNCTION get_pending_tasks_count(p_vds_id INTEGER)
RETURNS INTEGER AS $$
SELECT COUNT(*)::INTEGER
FROM tasks
WHERE vds_id = p_vds_id
  AND status IN ('pending', 'running');
$$ LANGUAGE SQL STABLE;

COMMENT ON FUNCTION get_pending_tasks_count IS 'Returns count of pending/running tasks for a VDS';

-- ============================================================================
-- STATISTICS (useful for monitoring)
-- ============================================================================

SELECT 'Database initialized successfully!' as message;
SELECT 'Plans:', COUNT(*) FROM plans;
SELECT 'Nodes:', COUNT(*) FROM nodes;
SELECT 'VDS instances:', COUNT(*) FROM vds;
SELECT 'Tasks:', COUNT(*) FROM tasks;
