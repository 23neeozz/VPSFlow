export type AuthTokens = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  mfa_required?: boolean;
  challenge_id?: string;
};

export type Organization = {
  org_id: string;
  name: string;
  slug: string;
  created_at?: string;
};

export type NodeCapacity = {
  cpu_cores: number;
  memory_bytes: number;
  storage_bytes: number;
  running_vms: number;
  allocated_cpu: number;
  allocated_memory_bytes: number;
  allocated_storage_bytes: number;
};

export type Hypervisor = {
  id: string;
  cluster_id: string;
  node_name: string;
  agent_version: string;
  status: string;
  maintenance_mode: boolean;
  last_heartbeat_at?: string;
  capacity: NodeCapacity;
};

export type VMStatus =
  | "pending"
  | "provisioning"
  | "running"
  | "stopping"
  | "stopped"
  | "starting"
  | "deleting"
  | "deleted"
  | "error";

export type VirtualMachine = {
  id: string;
  tenant_id: string;
  project_id?: string;
  name: string;
  status: VMStatus;
  hypervisor_id?: string;
  domain_name?: string;
  vcpus: number;
  memory_mb: number;
  disk_gb: number;
  image_ref?: string;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type VPSInstance = {
  id: string;
  tenant_id: string;
  vm_id: string;
  name: string;
  owner_user_id: string;
  owner_email: string;
  assigned_by_user_id: string;
  status: VMStatus;
  vcpus: number;
  memory_mb: number;
  disk_gb: number;
  error_message?: string;
  created_at: string;
  updated_at: string;
};

export type ConsoleSession = {
  session_id: string;
  token: string;
  proxy_url: string;
  expires_at: string;
};
