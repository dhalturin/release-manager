\connect "release-manager";

DROP TABLE IF EXISTS "repos";
DROP SEQUENCE repos_repo_id_seq;
CREATE SEQUENCE repos_repo_id_seq INCREMENT  MINVALUE  MAXVALUE  START 25 CACHE ;

CREATE TABLE "public"."repos" (
    "repo_id" integer DEFAULT nextval('repos_repo_id_seq') NOT NULL,
    "repo_name" text NOT NULL,
    "repo_ref" text NOT NULL,
    "repo_release" integer DEFAULT 0 NOT NULL,
    "repo_pipeline" integer DEFAULT 0 NOT NULL,
    "repo_prefix" text NOT NULL,
    "repo_channel" text NOT NULL,
    "repo_permission" text NOT NULL,
    "repo_admin" text NOT NULL,
    "repo_tasks" text DEFAULT  NOT NULL
) WITH (oids = false);

DROP TABLE IF EXISTS "users";
CREATE TABLE "public"."users" (
    "user_id" text NOT NULL,
    "user_token" text NOT NULL,
    "user_repo_name" text DEFAULT  NOT NULL,
    "user_repo_channel" text DEFAULT  NOT NULL,
    "user_repo_time" integer DEFAULT 0 NOT NULL
) WITH (oids = false);
