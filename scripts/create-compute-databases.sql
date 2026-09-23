-- BossCloud — bases de datos Compute (ejecutar como superusuario o usuario con CREATEDB)
-- Uso con psql:
--   "C:\Program Files\PostgreSQL\18\bin\psql.exe" -U postgres -h localhost -f scripts\create-compute-databases.sql

CREATE DATABASE bosscloud_cluster OWNER bosscloud;
CREATE DATABASE bosscloud_vm OWNER bosscloud;
CREATE DATABASE bosscloud_agent_control OWNER bosscloud;
CREATE DATABASE bosscloud_console OWNER bosscloud;
CREATE DATABASE bosscloud_vps OWNER bosscloud;
