-- One cluster, logical separation per service (no cross-DB queries from apps).
CREATE DATABASE velune_legacy;
CREATE DATABASE velune_auth;
CREATE DATABASE velune_transaction;
CREATE DATABASE velune_category;
CREATE DATABASE velune_budget;
CREATE DATABASE velune_report;
CREATE DATABASE velune_notification;
