CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE EXTENSION IF NOT EXISTS citext;

CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- =========================
-- ENUM TYPES
-- =========================
CREATE TYPE user_complexity AS ENUM ('beginner','intermediate','advanced');

CREATE TYPE project_visibility AS ENUM ('private','public','unlisted');

CREATE TYPE idea_status AS ENUM ('draft','saved','published','archived');

CREATE TYPE ai_interaction_type AS ENUM ('chat','code_suggestion','debug','convert','doc_generate');

CREATE TYPE vote_value AS ENUM ('up','down');

CREATE TYPE auth_provider AS ENUM ('email','google','github');

-- WARNING: This schema is for context only and is not meant to be run.
-- Table order and constraints may not be valid for execution.

CREATE TABLE public.accounts (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  email USER-DEFINED NOT NULL UNIQUE,
  email_verified boolean DEFAULT false,
  is_active boolean DEFAULT true,
  last_login timestamp with time zone,
  preferred_locale text,
  profile_metadata jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now(),
  CONSTRAINT accounts_pkey PRIMARY KEY (id)
);
CREATE TABLE public.ai_generation_logs (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  account_id uuid NOT NULL,
  project_id uuid,
  ai_model text NOT NULL,
  prompt_version text NOT NULL,
  input_metrics jsonb NOT NULL,
  output_result jsonb NOT NULL,
  created_at timestamp without time zone NOT NULL DEFAULT now(),
  CONSTRAINT ai_generation_logs_pkey PRIMARY KEY (id),
  CONSTRAINT ai_generation_logs_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id),
  CONSTRAINT ai_generation_logs_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id)
);
CREATE TABLE public.ai_prompt_templates (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  version text NOT NULL UNIQUE,
  description text,
  system_prompt text NOT NULL,
  generation_prompt text NOT NULL,
  output_format_prompt text NOT NULL,
  is_active boolean DEFAULT true,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT ai_prompt_templates_pkey PRIMARY KEY (id)
);
CREATE TABLE public.audit_logs (
  id bigint NOT NULL DEFAULT nextval('audit_logs_id_seq'::regclass),
  actor_account_id uuid,
  action text NOT NULL,
  target_type text,
  target_id uuid,
  details jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT audit_logs_pkey PRIMARY KEY (id)
);
CREATE TABLE public.comments (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  post_id uuid NOT NULL,
  account_id uuid,
  body text NOT NULL,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT comments_pkey PRIMARY KEY (id),
  CONSTRAINT comments_post_id_fkey FOREIGN KEY (post_id) REFERENCES public.posts(id),
  CONSTRAINT comments_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.documents (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  project_id uuid,
  account_id uuid,
  doc_type text,
  format text,
  content text,
  metadata jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT documents_pkey PRIMARY KEY (id),
  CONSTRAINT documents_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id),
  CONSTRAINT documents_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.files (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  owner_account_id uuid,
  project_id uuid,
  filename text NOT NULL,
  content_type text,
  size bigint,
  storage_path text,
  metadata jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT files_pkey PRIMARY KEY (id),
  CONSTRAINT files_owner_account_id_fkey FOREIGN KEY (owner_account_id) REFERENCES public.accounts(id),
  CONSTRAINT files_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id)
);
CREATE TABLE public.ideas (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  account_id uuid,
  title text,
  components ARRAY DEFAULT '{}'::text[],
  description text,
  feasibility jsonb DEFAULT '{}'::jsonb,
  resources jsonb DEFAULT '{}'::jsonb,
  status USER-DEFINED DEFAULT 'draft'::idea_status,
  saved boolean DEFAULT false,
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now(),
  CONSTRAINT ideas_pkey PRIMARY KEY (id),
  CONSTRAINT ideas_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.interactions (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  account_id uuid,
  project_id uuid,
  interaction_type USER-DEFINED NOT NULL,
  prompt text,
  response text,
  response_json jsonb DEFAULT '{}'::jsonb,
  model_version text,
  compute_cost numeric,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT interactions_pkey PRIMARY KEY (id),
  CONSTRAINT interactions_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id),
  CONSTRAINT interactions_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id)
);
CREATE TABLE public.local_credentials (
  account_id uuid NOT NULL,
  password_hash text NOT NULL,
  password_salt text,
  updated_at timestamp with time zone DEFAULT now(),
  CONSTRAINT local_credentials_pkey PRIMARY KEY (account_id),
  CONSTRAINT local_credentials_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.oauth_credentials (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  account_id uuid NOT NULL,
  provider USER-DEFINED NOT NULL,
  provider_user_id text,
  access_token text,
  refresh_token text,
  token_expires timestamp with time zone,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT oauth_credentials_pkey PRIMARY KEY (id),
  CONSTRAINT oauth_credentials_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.posts (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  account_id uuid,
  title text NOT NULL,
  body text NOT NULL,
  metadata jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now(),
  is_pinned boolean DEFAULT false,
  CONSTRAINT posts_pkey PRIMARY KEY (id),
  CONSTRAINT posts_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.project_dna_profiles (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  project_id uuid NOT NULL UNIQUE,
  domain text NOT NULL,
  target_user text NOT NULL,
  core_problem text NOT NULL,
  feature_signals ARRAY NOT NULL,
  architecture_style text NOT NULL,
  data_model_complexity text CHECK (data_model_complexity = ANY (ARRAY['Low'::text, 'Medium'::text, 'High'::text])),
  non_functional_constraints ARRAY NOT NULL,
  tech_stack ARRAY NOT NULL,
  dna_signature text NOT NULL,
  dna_hash text NOT NULL,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT project_dna_profiles_pkey PRIMARY KEY (id),
  CONSTRAINT project_dna_profiles_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id)
);
CREATE TABLE public.project_dna_registry (
  dna_hash text NOT NULL,
  first_project_id uuid,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT project_dna_registry_pkey PRIMARY KEY (dna_hash),
  CONSTRAINT project_dna_registry_first_project_id_fkey FOREIGN KEY (first_project_id) REFERENCES public.projects(id)
);
CREATE TABLE public.project_generation_attempts (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  account_id uuid NOT NULL,
  project_id uuid,
  attempt_number integer NOT NULL,
  generation_status text NOT NULL CHECK (generation_status = ANY (ARRAY['generated'::text, 'pivoted'::text, 'rejected'::text, 'accepted'::text])),
  pivot_reason text,
  forbidden_reuse ARRAY,
  pivot_strategy ARRAY,
  ai_model text NOT NULL,
  prompt_version text NOT NULL,
  created_at timestamp without time zone NOT NULL DEFAULT now(),
  CONSTRAINT project_generation_attempts_pkey PRIMARY KEY (id),
  CONSTRAINT project_generation_attempts_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id),
  CONSTRAINT project_generation_attempts_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id)
);
CREATE TABLE public.project_similarity_checks (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  project_id uuid NOT NULL,
  compared_project_id uuid NOT NULL,
  similarity_score numeric NOT NULL,
  similarity_threshold numeric NOT NULL DEFAULT 0.75,
  is_duplicate boolean NOT NULL,
  checked_dimensions ARRAY NOT NULL,
  created_at timestamp without time zone NOT NULL DEFAULT now(),
  CONSTRAINT project_similarity_checks_pkey PRIMARY KEY (id),
  CONSTRAINT project_similarity_checks_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id),
  CONSTRAINT project_similarity_checks_compared_project_id_fkey FOREIGN KEY (compared_project_id) REFERENCES public.projects(id)
);
CREATE TABLE public.project_tags (
  project_id uuid NOT NULL,
  tag_id uuid NOT NULL,
  CONSTRAINT project_tags_pkey PRIMARY KEY (project_id, tag_id),
  CONSTRAINT project_tags_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id),
  CONSTRAINT project_tags_tag_id_fkey FOREIGN KEY (tag_id) REFERENCES public.tags(id)
);
CREATE TABLE public.project_technologies (
  project_id uuid NOT NULL,
  technology_id uuid NOT NULL,
  importance smallint DEFAULT 1,
  CONSTRAINT project_technologies_pkey PRIMARY KEY (project_id, technology_id),
  CONSTRAINT project_technologies_project_id_fkey FOREIGN KEY (project_id) REFERENCES public.projects(id),
  CONSTRAINT project_technologies_technology_id_fkey FOREIGN KEY (technology_id) REFERENCES public.technologies(id)
);
CREATE TABLE public.projects (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  title text NOT NULL,
  slug text NOT NULL UNIQUE,
  short_description text,
  long_description text,
  creator_account_id uuid,
  complexity USER-DEFINED DEFAULT 'beginner'::user_complexity,
  estimated_duration text,
  visibility USER-DEFINED DEFAULT 'public'::project_visibility,
  metadata jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now(),
  project_dna_hash text,
  generation_version text,
  uniqueness_score numeric,
  last_similarity_check_at timestamp without time zone,
  CONSTRAINT projects_pkey PRIMARY KEY (id),
  CONSTRAINT projects_creator_account_id_fkey FOREIGN KEY (creator_account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.tags (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT tags_pkey PRIMARY KEY (id)
);
CREATE TABLE public.technologies (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  name text NOT NULL,
  slug text NOT NULL UNIQUE,
  category text,
  meta jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT technologies_pkey PRIMARY KEY (id)
);
CREATE TABLE public.trending_tags (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  tag text NOT NULL,
  source text NOT NULL,
  period_date date NOT NULL,
  score numeric,
  metadata jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT trending_tags_pkey PRIMARY KEY (id)
);
CREATE TABLE public.user_profiles (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  account_id uuid NOT NULL UNIQUE,
  display_name text,
  avatar_url text,
  bio text,
  skills ARRAY DEFAULT '{}'::text[],
  favorite_languages ARRAY DEFAULT '{}'::text[],
  complexity_pref USER-DEFINED DEFAULT 'beginner'::user_complexity,
  settings jsonb DEFAULT '{}'::jsonb,
  created_at timestamp with time zone DEFAULT now(),
  updated_at timestamp with time zone DEFAULT now(),
  CONSTRAINT user_profiles_pkey PRIMARY KEY (id),
  CONSTRAINT user_profiles_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);
CREATE TABLE public.votes (
  id uuid NOT NULL DEFAULT uuid_generate_v4(),
  target_type text NOT NULL,
  target_id uuid NOT NULL,
  account_id uuid,
  value USER-DEFINED NOT NULL,
  created_at timestamp with time zone DEFAULT now(),
  CONSTRAINT votes_pkey PRIMARY KEY (id),
  CONSTRAINT votes_account_id_fkey FOREIGN KEY (account_id) REFERENCES public.accounts(id)
);