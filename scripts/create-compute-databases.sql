-- VPSFlow — bases de datos Compute (ejecutar como superusuario o usuario con CREATEDB)
-- Uso con psql:
--   "C:\Program Files\PostgreSQL\18\bin\psql.exe" -U postgres -h localhost -f scripts\create-compute-databases.sql

CREATE DATABASE vpsflow_cluster OWNER vpsflow;
CREATE DATABASE vpsflow_vm OWNER vpsflow;
CREATE DATABASE vpsflow_agent_control OWNER vpsflow;
CREATE DATABASE vpsflow_console OWNER vpsflow;
CREATE DATABASE vpsflow_vps OWNER vpsflow;
