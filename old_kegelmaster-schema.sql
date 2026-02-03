--
-- PostgreSQL database dump
--

\restrict AFqcUbbLzJDNkRQw0eYdMOR8cbxDNvzA6BVVLUsm6gGgiIIsBBL6r2qeDYP5wCA

-- Dumped from database version 17.7 (Debian 17.7-3.pgdg13+1)
-- Dumped by pg_dump version 18.1

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: public; Type: SCHEMA; Schema: -; Owner: -
--

CREATE SCHEMA public;


--
-- Name: SCHEMA public; Type: COMMENT; Schema: -; Owner: -
--

COMMENT ON SCHEMA public IS 'standard public schema';


SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: cache; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cache (
    key character varying(255) NOT NULL,
    value text NOT NULL,
    expiration integer NOT NULL
);


--
-- Name: cache_locks; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.cache_locks (
    key character varying(255) NOT NULL,
    owner character varying(255) NOT NULL,
    expiration integer NOT NULL
);


--
-- Name: club_settings; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.club_settings (
    id bigint NOT NULL,
    club_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    value integer NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: club_settings_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.club_settings_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: club_settings_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.club_settings_id_seq OWNED BY public.club_settings.id;


--
-- Name: clubs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.clubs (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    balance integer NOT NULL,
    base_fee integer DEFAULT 0 NOT NULL,
    initial_balance integer DEFAULT 0 NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: clubs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.clubs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: clubs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.clubs_id_seq OWNED BY public.clubs.id;


--
-- Name: competition_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.competition_entries (
    id bigint NOT NULL,
    matchday_id bigint NOT NULL,
    player_id bigint NOT NULL,
    competition_type_id bigint NOT NULL,
    amount integer NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: competition_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.competition_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: competition_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.competition_entries_id_seq OWNED BY public.competition_entries.id;


--
-- Name: competition_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.competition_types (
    id bigint NOT NULL,
    club_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    type smallint NOT NULL,
    is_sex_specific boolean NOT NULL,
    "position" smallint DEFAULT '0'::smallint NOT NULL,
    description character varying(255),
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: competition_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.competition_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: competition_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.competition_types_id_seq OWNED BY public.competition_types.id;


--
-- Name: dashboard_layouts; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.dashboard_layouts (
    id bigint NOT NULL,
    user_id bigint NOT NULL,
    club_id bigint NOT NULL,
    layout json NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: dashboard_layouts_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.dashboard_layouts_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: dashboard_layouts_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.dashboard_layouts_id_seq OWNED BY public.dashboard_layouts.id;


--
-- Name: failed_jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.failed_jobs (
    id bigint NOT NULL,
    uuid character varying(255) NOT NULL,
    connection text NOT NULL,
    queue text NOT NULL,
    payload text NOT NULL,
    exception text NOT NULL,
    failed_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP NOT NULL
);


--
-- Name: failed_jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.failed_jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: failed_jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.failed_jobs_id_seq OWNED BY public.failed_jobs.id;


--
-- Name: fee_entries; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fee_entries (
    id bigint NOT NULL,
    matchday_id bigint NOT NULL,
    player_id bigint NOT NULL,
    fee_type_version_id bigint NOT NULL,
    amount double precision DEFAULT '0'::double precision NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: fee_entries_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fee_entries_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fee_entries_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fee_entries_id_seq OWNED BY public.fee_entries.id;


--
-- Name: fee_type_version_matchday; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fee_type_version_matchday (
    fee_type_version_id bigint NOT NULL,
    matchday_id bigint NOT NULL
);


--
-- Name: fee_type_versions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fee_type_versions (
    id bigint NOT NULL,
    fee_type_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    description text,
    amount integer NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: fee_type_versions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fee_type_versions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fee_type_versions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fee_type_versions_id_seq OWNED BY public.fee_type_versions.id;


--
-- Name: fee_types; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.fee_types (
    id bigint NOT NULL,
    club_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    description character varying(255),
    amount integer DEFAULT 0 NOT NULL,
    "position" smallint DEFAULT '0'::smallint NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: fee_types_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.fee_types_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: fee_types_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.fee_types_id_seq OWNED BY public.fee_types.id;


--
-- Name: job_batches; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.job_batches (
    id character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    total_jobs integer NOT NULL,
    pending_jobs integer NOT NULL,
    failed_jobs integer NOT NULL,
    failed_job_ids text NOT NULL,
    options text,
    cancelled_at integer,
    created_at integer NOT NULL,
    finished_at integer
);


--
-- Name: jobs; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.jobs (
    id bigint NOT NULL,
    queue character varying(255) NOT NULL,
    payload text NOT NULL,
    attempts smallint NOT NULL,
    reserved_at integer,
    available_at integer NOT NULL,
    created_at integer NOT NULL
);


--
-- Name: jobs_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.jobs_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: jobs_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.jobs_id_seq OWNED BY public.jobs.id;


--
-- Name: matchday_player; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.matchday_player (
    matchday_id bigint NOT NULL,
    player_id bigint NOT NULL,
    created_at timestamp(0) without time zone DEFAULT CURRENT_TIMESTAMP
);


--
-- Name: matchdays; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.matchdays (
    id bigint NOT NULL,
    club_id bigint NOT NULL,
    date date NOT NULL,
    notes text,
    is_calculated boolean DEFAULT false NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: matchdays_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.matchdays_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: matchdays_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.matchdays_id_seq OWNED BY public.matchdays.id;


--
-- Name: migrations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.migrations (
    id integer NOT NULL,
    migration character varying(255) NOT NULL,
    batch integer NOT NULL
);


--
-- Name: migrations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.migrations_id_seq
    AS integer
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: migrations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.migrations_id_seq OWNED BY public.migrations.id;


--
-- Name: model_has_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.model_has_permissions (
    permission_id bigint NOT NULL,
    model_type character varying(255) NOT NULL,
    model_id bigint NOT NULL,
    club_id bigint NOT NULL
);


--
-- Name: model_has_roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.model_has_roles (
    role_id bigint NOT NULL,
    model_type character varying(255) NOT NULL,
    model_id bigint NOT NULL,
    club_id bigint NOT NULL
);


--
-- Name: password_reset_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.password_reset_tokens (
    email character varying(255) NOT NULL,
    token character varying(255) NOT NULL,
    created_at timestamp(0) without time zone
);


--
-- Name: permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.permissions (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    guard_name character varying(255) NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: permissions_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.permissions_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: permissions_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.permissions_id_seq1 OWNED BY public.permissions.id;


--
-- Name: personal_access_tokens; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.personal_access_tokens (
    id bigint NOT NULL,
    tokenable_type character varying(255) NOT NULL,
    tokenable_id bigint NOT NULL,
    name character varying(255) NOT NULL,
    token character varying(64) NOT NULL,
    abilities text,
    last_used_at timestamp(0) without time zone,
    expires_at timestamp(0) without time zone,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: personal_access_tokens_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.personal_access_tokens_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: personal_access_tokens_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.personal_access_tokens_id_seq OWNED BY public.personal_access_tokens.id;


--
-- Name: player_invitations; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.player_invitations (
    id bigint NOT NULL,
    player_id bigint NOT NULL,
    email character varying(255) NOT NULL,
    token character varying(255) NOT NULL,
    expires_at timestamp(0) without time zone NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: player_invitations_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.player_invitations_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: player_invitations_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.player_invitations_id_seq OWNED BY public.player_invitations.id;


--
-- Name: players; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.players (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    club_id bigint NOT NULL,
    user_id bigint,
    sex integer DEFAULT 0 NOT NULL,
    balance integer DEFAULT 0 NOT NULL,
    initial_balance integer DEFAULT 0 NOT NULL,
    active boolean DEFAULT true NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    role_id bigint NOT NULL
);


--
-- Name: players_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.players_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: players_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.players_id_seq OWNED BY public.players.id;


--
-- Name: role_has_permissions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.role_has_permissions (
    permission_id bigint NOT NULL,
    role_id bigint NOT NULL
);


--
-- Name: roles; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.roles (
    id bigint NOT NULL,
    club_id bigint,
    name character varying(255) NOT NULL,
    guard_name character varying(255) NOT NULL,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone,
    is_base_fee_active boolean DEFAULT false NOT NULL
);


--
-- Name: roles_id_seq1; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.roles_id_seq1
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: roles_id_seq1; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.roles_id_seq1 OWNED BY public.roles.id;


--
-- Name: sessions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.sessions (
    id character varying(255) NOT NULL,
    user_id bigint,
    ip_address character varying(45),
    user_agent text,
    payload text NOT NULL,
    last_activity integer NOT NULL
);


--
-- Name: transactions; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.transactions (
    id bigint NOT NULL,
    club_id bigint,
    player_id bigint,
    matchday_id bigint,
    fee_entry_id bigint,
    type integer NOT NULL,
    amount integer NOT NULL,
    date date NOT NULL,
    notes text,
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: transactions_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.transactions_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: transactions_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.transactions_id_seq OWNED BY public.transactions.id;


--
-- Name: users; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.users (
    id bigint NOT NULL,
    name character varying(255) NOT NULL,
    email character varying(255) NOT NULL,
    email_verified_at timestamp(0) without time zone,
    password character varying(255) NOT NULL,
    remember_token character varying(100),
    created_at timestamp(0) without time zone,
    updated_at timestamp(0) without time zone
);


--
-- Name: users_id_seq; Type: SEQUENCE; Schema: public; Owner: -
--

CREATE SEQUENCE public.users_id_seq
    START WITH 1
    INCREMENT BY 1
    NO MINVALUE
    NO MAXVALUE
    CACHE 1;


--
-- Name: users_id_seq; Type: SEQUENCE OWNED BY; Schema: public; Owner: -
--

ALTER SEQUENCE public.users_id_seq OWNED BY public.users.id;


--
-- Name: club_settings id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.club_settings ALTER COLUMN id SET DEFAULT nextval('public.club_settings_id_seq'::regclass);


--
-- Name: clubs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clubs ALTER COLUMN id SET DEFAULT nextval('public.clubs_id_seq'::regclass);


--
-- Name: competition_entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_entries ALTER COLUMN id SET DEFAULT nextval('public.competition_entries_id_seq'::regclass);


--
-- Name: competition_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_types ALTER COLUMN id SET DEFAULT nextval('public.competition_types_id_seq'::regclass);


--
-- Name: dashboard_layouts id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dashboard_layouts ALTER COLUMN id SET DEFAULT nextval('public.dashboard_layouts_id_seq'::regclass);


--
-- Name: failed_jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_jobs ALTER COLUMN id SET DEFAULT nextval('public.failed_jobs_id_seq'::regclass);


--
-- Name: fee_entries id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_entries ALTER COLUMN id SET DEFAULT nextval('public.fee_entries_id_seq'::regclass);


--
-- Name: fee_type_versions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_type_versions ALTER COLUMN id SET DEFAULT nextval('public.fee_type_versions_id_seq'::regclass);


--
-- Name: fee_types id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_types ALTER COLUMN id SET DEFAULT nextval('public.fee_types_id_seq'::regclass);


--
-- Name: jobs id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs ALTER COLUMN id SET DEFAULT nextval('public.jobs_id_seq'::regclass);


--
-- Name: matchdays id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matchdays ALTER COLUMN id SET DEFAULT nextval('public.matchdays_id_seq'::regclass);


--
-- Name: migrations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.migrations ALTER COLUMN id SET DEFAULT nextval('public.migrations_id_seq'::regclass);


--
-- Name: permissions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions ALTER COLUMN id SET DEFAULT nextval('public.permissions_id_seq1'::regclass);


--
-- Name: personal_access_tokens id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.personal_access_tokens ALTER COLUMN id SET DEFAULT nextval('public.personal_access_tokens_id_seq'::regclass);


--
-- Name: player_invitations id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.player_invitations ALTER COLUMN id SET DEFAULT nextval('public.player_invitations_id_seq'::regclass);


--
-- Name: players id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.players ALTER COLUMN id SET DEFAULT nextval('public.players_id_seq'::regclass);


--
-- Name: roles id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles ALTER COLUMN id SET DEFAULT nextval('public.roles_id_seq1'::regclass);


--
-- Name: transactions id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions ALTER COLUMN id SET DEFAULT nextval('public.transactions_id_seq'::regclass);


--
-- Name: users id; Type: DEFAULT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users ALTER COLUMN id SET DEFAULT nextval('public.users_id_seq'::regclass);


--
-- Data for Name: cache; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.cache (key, value, expiration) FROM stdin;
\.


--
-- Data for Name: cache_locks; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.cache_locks (key, owner, expiration) FROM stdin;
\.


--
-- Data for Name: club_settings; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.club_settings (id, club_id, name, value, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: clubs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.clubs (id, user_id, name, balance, base_fee, initial_balance, created_at, updated_at) FROM stdin;
6	1	Test Club	0	0	0	2025-03-20 13:56:17	2025-03-20 13:56:17
1	1	Krumme Kugel	191220	1000	0	2025-03-08 11:34:32	2026-01-23 15:32:54
2	2	Die Legionäre	0	0	0	2025-03-09 10:31:30	2025-03-24 17:37:10
\.


--
-- Data for Name: competition_entries; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.competition_entries (id, matchday_id, player_id, competition_type_id, amount, created_at, updated_at) FROM stdin;
119	24	4	1	50	\N	2025-03-18 20:26:37
120	24	7	1	140	\N	2025-03-18 20:26:46
121	24	8	1	157	\N	2025-03-18 20:26:53
122	24	10	1	100	\N	2025-03-18 20:26:58
123	24	9	1	79	\N	2025-03-18 20:27:03
118	24	2	1	153	\N	2025-03-18 20:27:10
117	24	1	1	118	\N	2025-03-18 20:27:17
116	24	3	1	214	\N	2025-03-18 20:27:22
61	16	7	1	199	\N	2025-03-17 14:04:24
62	16	8	1	221	\N	2025-03-17 14:04:39
63	16	4	1	102	\N	2025-03-17 14:04:51
64	16	9	1	17	\N	2025-03-17 14:05:06
65	16	2	1	133	\N	2025-03-17 14:05:18
66	16	3	1	219	\N	2025-03-17 14:05:26
70	17	4	1	68	\N	2025-03-17 14:18:33
74	17	7	1	114	\N	2025-03-17 14:18:41
68	17	8	1	141	\N	2025-03-17 14:18:51
69	17	12	1	173	\N	2025-03-17 14:19:09
67	17	10	1	204	\N	2025-03-17 14:19:18
71	17	9	1	71	\N	2025-03-17 14:19:29
73	17	2	1	150	\N	2025-03-17 14:19:40
75	17	1	1	115	\N	2025-03-17 14:19:51
72	17	13	1	116	\N	2025-03-17 14:20:01
76	17	3	1	189	\N	2025-03-17 14:20:09
77	18	3	1	204	\N	2025-03-18 07:05:55
78	18	8	1	116	\N	2025-03-18 07:06:22
79	18	4	1	98	\N	2025-03-18 07:08:26
84	18	10	1	197	\N	2025-03-18 07:08:32
80	18	9	1	87	\N	2025-03-18 07:08:38
82	18	2	1	109	\N	2025-03-18 07:08:43
83	18	1	1	158	\N	2025-03-18 07:08:49
81	18	11	1	128	\N	2025-03-18 07:08:56
92	20	4	1	106	\N	2025-03-18 07:14:04
91	20	8	1	165	\N	2025-03-18 07:14:09
90	20	10	1	161	\N	2025-03-18 07:14:15
94	20	2	1	83	\N	2025-03-18 07:14:20
93	20	11	1	68	\N	2025-03-18 07:14:26
95	20	3	1	194	\N	2025-03-18 07:14:33
86	19	7	1	151	\N	2025-03-18 07:14:50
85	19	10	1	171	\N	2025-03-18 07:14:56
88	19	2	1	133	\N	2025-03-18 07:15:01
89	19	1	1	138	\N	2025-03-18 07:15:05
87	19	11	1	131	\N	2025-03-18 07:15:13
100	21	7	1	164	\N	2025-03-18 07:17:28
96	21	10	1	221	\N	2025-03-18 07:17:41
98	21	2	1	153	\N	2025-03-18 07:17:49
99	21	1	1	118	\N	2025-03-18 07:17:53
97	21	11	1	78	\N	2025-03-18 07:17:57
103	22	14	1	102	\N	2025-03-18 07:21:06
106	22	5	1	184	\N	2025-03-18 07:21:11
101	22	10	1	192	\N	2025-03-18 07:21:18
102	22	9	1	75	\N	2025-03-18 07:21:27
104	22	2	1	179	\N	2025-03-18 07:21:37
105	22	1	1	158	\N	2025-03-18 07:21:47
107	23	4	1	62	\N	2025-03-18 20:22:13
114	23	7	1	124	\N	2025-03-18 20:22:22
108	23	8	1	171	\N	2025-03-18 20:22:29
115	23	12	1	197	\N	2025-03-18 20:22:37
111	23	10	1	208	\N	2025-03-18 20:22:43
113	23	9	1	64	\N	2025-03-18 20:22:49
110	23	11	1	105	\N	2025-03-18 20:22:56
112	23	13	1	80	\N	2025-03-18 20:23:03
109	23	3	1	246	\N	2025-03-18 20:23:09
142	31	4	1	59	\N	2025-03-21 16:42:21
143	31	8	1	137	\N	2025-03-21 16:42:27
144	31	12	1	197	\N	2025-03-21 16:42:36
151	31	10	1	150	\N	2025-03-21 16:42:43
145	31	9	1	104	\N	2025-03-21 16:42:49
146	31	2	1	126	\N	2025-03-21 16:42:57
147	31	1	1	140	\N	2025-03-21 16:43:03
148	31	11	1	67	\N	2025-03-21 16:43:07
149	31	13	1	66	\N	2025-03-21 16:43:13
150	31	3	1	226	\N	2025-03-21 16:43:20
182	34	30	3	168	\N	2025-03-25 12:36:05
183	34	30	4	0	\N	2025-03-25 12:36:05
192	34	34	3	175	\N	2025-03-25 12:36:47
193	34	34	4	0	\N	2025-03-25 12:36:47
185	34	36	4	0	\N	2025-03-27 15:34:59
186	34	32	3	191	\N	2025-03-27 15:35:10
187	34	32	4	0	\N	2025-03-27 15:35:11
188	34	33	3	190	\N	2025-03-27 15:35:49
189	34	33	4	0	\N	2025-03-27 15:35:50
191	34	35	4	0	\N	2025-03-27 15:36:11
194	34	37	3	243	\N	2025-03-27 15:36:24
195	34	37	4	0	\N	2025-03-27 15:36:24
197	35	30	4	0	\N	2025-03-27 15:37:47
200	35	32	3	131	\N	2025-03-27 15:38:31
199	35	36	4	0	\N	2025-03-27 15:37:42
196	35	30	3	163	\N	2025-03-27 15:37:47
201	35	32	4	0	\N	2025-03-27 15:38:32
206	35	33	3	191	\N	2025-03-27 15:38:45
207	35	33	4	0	\N	2025-03-27 15:38:45
203	35	34	4	0	\N	2025-03-27 15:39:24
204	35	35	3	138	\N	2025-03-27 15:39:35
205	35	35	4	0	\N	2025-03-27 15:39:35
208	35	37	3	225	\N	2025-03-27 15:40:14
209	35	37	4	0	\N	2025-03-27 15:40:14
210	35	31	3	158	\N	2025-03-27 15:40:26
212	36	36	3	195	\N	2025-04-25 12:15:42
214	36	37	3	214	\N	2025-04-25 12:16:10
215	36	37	4	199	\N	2025-04-25 12:16:11
213	36	36	4	202	\N	2025-04-25 12:16:31
184	34	36	3	193	\N	2025-03-27 15:34:59
190	34	35	3	90	\N	2025-03-27 15:36:10
198	35	36	3	168	\N	2025-03-27 15:37:42
202	35	34	3	177	\N	2025-03-27 15:39:24
211	35	31	4	0	\N	2025-03-27 15:40:26
220	45	36	3	199	2025-04-25 12:06:24	2025-04-25 12:06:24
221	45	36	4	0	2025-04-25 12:06:24	2025-04-25 12:06:24
222	45	35	3	139	2025-04-25 12:06:42	2025-04-25 12:06:42
223	45	35	4	0	2025-04-25 12:06:42	2025-04-25 12:06:42
224	45	32	3	128	2025-04-25 12:06:56	2025-04-25 12:06:56
225	45	32	4	0	2025-04-25 12:06:56	2025-04-25 12:06:56
226	45	33	3	98	2025-04-25 12:07:03	2025-04-25 12:07:03
227	45	33	4	0	2025-04-25 12:07:03	2025-04-25 12:07:03
228	45	34	3	188	2025-04-25 12:07:13	2025-04-25 12:07:13
229	45	34	4	0	2025-04-25 12:07:13	2025-04-25 12:07:13
230	45	37	3	209	2025-04-25 12:07:20	2025-04-25 12:07:20
231	45	37	4	0	2025-04-25 12:07:20	2025-04-25 12:07:20
232	45	31	3	187	2025-04-25 12:07:28	2025-04-25 12:07:28
233	45	31	4	0	2025-04-25 12:07:28	2025-04-25 12:07:28
234	44	36	3	176	2025-04-25 12:10:41	2025-04-25 12:10:41
235	44	36	4	0	2025-04-25 12:10:41	2025-04-25 12:10:41
236	44	32	3	204	2025-04-25 12:10:50	2025-04-25 12:10:50
237	44	32	4	0	2025-04-25 12:10:51	2025-04-25 12:10:51
238	44	35	3	133	2025-04-25 12:10:59	2025-04-25 12:10:59
239	44	35	4	0	2025-04-25 12:10:59	2025-04-25 12:10:59
240	44	37	3	199	2025-04-25 12:11:06	2025-04-25 12:11:06
241	44	37	4	0	2025-04-25 12:11:06	2025-04-25 12:11:06
242	44	31	3	198	2025-04-25 12:11:14	2025-04-25 12:11:14
243	44	31	4	0	2025-04-25 12:11:14	2025-04-25 12:11:14
218	36	33	3	186	\N	2025-04-25 12:16:01
219	36	33	4	151	\N	2025-04-25 12:16:02
216	36	31	3	165	\N	2025-04-25 12:16:23
217	36	31	4	206	\N	2025-04-25 12:16:23
244	47	30	3	78	2025-04-25 12:21:33	2025-04-25 12:21:33
245	47	30	4	0	2025-04-25 12:21:33	2025-04-25 12:21:33
246	47	36	3	227	2025-04-25 12:21:38	2025-04-25 12:21:38
247	47	36	4	0	2025-04-25 12:21:38	2025-04-25 12:21:38
248	47	32	3	155	2025-04-25 12:21:44	2025-04-25 12:21:44
249	47	32	4	0	2025-04-25 12:21:44	2025-04-25 12:21:44
250	47	33	3	147	2025-04-25 12:21:50	2025-04-25 12:21:50
251	47	33	4	0	2025-04-25 12:21:50	2025-04-25 12:21:50
252	47	34	3	79	2025-04-25 12:21:55	2025-04-25 12:21:55
253	47	34	4	0	2025-04-25 12:21:55	2025-04-25 12:21:55
254	47	35	3	188	2025-04-25 12:22:03	2025-04-25 12:22:03
255	47	35	4	0	2025-04-25 12:22:03	2025-04-25 12:22:03
256	47	31	3	186	2025-04-25 12:22:08	2025-04-25 12:22:08
257	47	31	4	0	2025-04-25 12:22:08	2025-04-25 12:22:08
258	40	30	3	158	2025-04-25 12:23:35	2025-04-25 12:23:35
259	40	30	4	0	2025-04-25 12:23:35	2025-04-25 12:23:35
261	40	36	4	0	2025-04-25 12:23:40	2025-04-25 12:23:40
262	40	33	3	201	2025-04-25 12:23:47	2025-04-25 12:23:47
263	40	33	4	0	2025-04-25 12:23:48	2025-04-25 12:23:48
260	40	36	3	178	2025-04-25 12:23:40	2025-04-25 12:24:03
264	40	35	3	161	2025-04-25 12:24:10	2025-04-25 12:24:10
265	40	35	4	0	2025-04-25 12:24:10	2025-04-25 12:24:10
266	40	37	3	218	2025-04-25 12:24:18	2025-04-25 12:24:18
267	40	37	4	0	2025-04-25 12:24:18	2025-04-25 12:24:18
268	40	31	3	174	2025-04-25 12:24:24	2025-04-25 12:24:24
269	40	31	4	0	2025-04-25 12:24:25	2025-04-25 12:24:25
270	48	30	3	126	2025-04-25 12:26:51	2025-04-25 12:26:51
271	48	30	4	127	2025-04-25 12:26:51	2025-04-25 12:26:51
272	48	32	3	114	2025-04-25 12:27:11	2025-04-25 12:27:11
273	48	32	4	148	2025-04-25 12:27:11	2025-04-25 12:27:11
276	48	34	3	201	2025-04-25 12:27:27	2025-04-25 12:27:27
277	48	34	4	159	2025-04-25 12:27:28	2025-04-25 12:27:28
278	48	31	3	216	2025-04-25 12:27:39	2025-04-25 12:27:39
279	48	31	4	139	2025-04-25 12:27:40	2025-04-25 12:27:40
274	48	33	3	201	2025-04-25 12:27:19	2025-04-25 12:27:47
275	48	33	4	193	2025-04-25 12:27:19	2025-04-25 12:27:47
280	42	32	3	119	2025-04-25 12:28:51	2025-04-25 12:28:51
281	42	32	4	158	2025-04-25 12:28:51	2025-04-25 12:28:51
282	42	34	3	216	2025-04-25 12:29:04	2025-04-25 12:29:04
283	42	34	4	225	2025-04-25 12:29:04	2025-04-25 12:29:04
284	42	35	3	115	2025-04-25 12:29:12	2025-04-25 12:29:12
285	42	35	4	136	2025-04-25 12:29:13	2025-04-25 12:29:13
286	42	37	3	184	2025-04-25 12:29:19	2025-04-25 12:29:19
287	42	37	4	205	2025-04-25 12:29:19	2025-04-25 12:29:19
288	43	30	3	138	2025-04-25 12:32:19	2025-04-25 12:32:19
289	43	30	4	0	2025-04-25 12:32:19	2025-04-25 12:32:19
290	43	36	3	219	2025-04-25 12:32:26	2025-04-25 12:32:26
291	43	36	4	0	2025-04-25 12:32:26	2025-04-25 12:32:26
292	43	32	3	194	2025-04-25 12:32:33	2025-04-25 12:32:33
293	43	32	4	0	2025-04-25 12:32:33	2025-04-25 12:32:33
294	43	33	3	225	2025-04-25 12:32:39	2025-04-25 12:32:39
295	43	33	4	0	2025-04-25 12:32:39	2025-04-25 12:32:39
296	43	34	3	222	2025-04-25 12:32:46	2025-04-25 12:32:46
297	43	34	4	0	2025-04-25 12:32:46	2025-04-25 12:32:46
298	43	35	3	155	2025-04-25 12:32:54	2025-04-25 12:32:54
299	43	35	4	0	2025-04-25 12:32:55	2025-04-25 12:32:55
300	43	37	3	192	2025-04-25 12:33:01	2025-04-25 12:33:01
301	43	37	4	0	2025-04-25 12:33:01	2025-04-25 12:33:01
302	43	31	3	167	2025-04-25 12:33:15	2025-04-25 12:33:15
303	43	31	4	0	2025-04-25 12:33:15	2025-04-25 12:33:15
315	50	36	4	0	2025-04-25 12:56:39	2025-04-25 12:56:39
316	50	32	3	204	2025-04-25 12:56:47	2025-04-25 12:56:47
317	50	32	4	0	2025-04-25 12:56:47	2025-04-25 12:56:47
314	50	36	3	176	2025-04-25 12:56:39	2025-04-25 12:56:54
318	50	35	3	133	2025-04-25 12:57:02	2025-04-25 12:57:02
319	50	35	4	0	2025-04-25 12:57:02	2025-04-25 12:57:02
320	50	37	3	199	2025-04-25 12:57:13	2025-04-25 12:57:13
321	50	37	4	0	2025-04-25 12:57:13	2025-04-25 12:57:13
322	50	31	3	198	2025-04-25 12:57:19	2025-04-25 12:57:19
323	50	31	4	0	2025-04-25 12:57:19	2025-04-25 12:57:19
324	54	4	1	71	2025-06-13 15:42:15	2025-06-13 15:42:15
325	54	7	1	154	2025-06-13 15:42:21	2025-06-13 15:42:21
326	54	8	1	160	2025-06-13 15:42:29	2025-06-13 15:42:29
327	54	10	1	215	2025-06-13 15:42:35	2025-06-13 15:42:35
328	54	9	1	58	2025-06-13 15:42:41	2025-06-13 15:42:41
329	54	2	1	134	2025-06-13 15:42:50	2025-06-13 15:42:50
330	54	1	1	151	2025-06-13 15:42:56	2025-06-13 15:42:56
331	54	11	1	124	2025-06-13 15:43:04	2025-06-13 15:43:04
332	54	3	1	242	2025-06-13 15:43:10	2025-06-13 15:43:10
333	55	4	1	71	2025-08-06 10:39:21	2025-08-06 10:39:21
334	55	8	1	182	2025-08-06 10:39:35	2025-08-06 10:39:35
335	55	10	1	180	2025-08-06 10:39:53	2025-08-06 10:39:53
336	55	9	1	62	2025-08-06 10:40:03	2025-08-06 10:40:03
337	55	2	1	174	2025-08-06 10:40:18	2025-08-06 10:40:18
338	55	1	1	131	2025-08-06 10:40:28	2025-08-06 10:40:28
339	55	11	1	104	2025-08-06 10:40:52	2025-08-06 10:40:52
340	55	3	1	148	2025-08-06 10:41:02	2025-08-06 10:41:02
341	56	7	1	141	2025-08-06 10:44:58	2025-08-06 10:44:58
342	56	12	1	180	2025-08-06 10:45:07	2025-08-06 10:45:07
343	56	2	1	133	2025-08-06 10:45:24	2025-08-06 10:45:24
344	56	1	1	109	2025-08-06 10:45:34	2025-08-06 10:45:34
345	56	13	1	124	2025-08-06 10:45:44	2025-08-06 10:45:44
346	57	10	1	189	2025-09-05 15:16:26	2025-09-05 15:16:26
347	57	8	1	97	2025-09-05 15:16:34	2025-09-05 15:16:34
348	57	9	1	66	2025-09-05 15:16:40	2025-09-05 15:16:40
349	57	11	1	134	2025-09-05 15:16:47	2025-09-05 15:16:47
350	57	2	1	142	2025-09-05 15:16:55	2025-09-05 15:16:55
351	57	1	1	117	2025-09-05 15:17:03	2025-09-05 15:17:03
352	57	7	1	118	2025-09-05 15:17:11	2025-09-05 15:17:11
353	58	10	1	200	2025-09-11 12:59:22	2025-09-11 12:59:22
354	58	8	1	142	2025-09-11 12:59:31	2025-09-11 12:59:31
355	58	4	1	59	2025-09-11 12:59:40	2025-09-11 12:59:40
356	58	9	1	109	2025-09-11 12:59:47	2025-09-11 12:59:47
357	58	11	1	121	2025-09-11 12:59:54	2025-09-11 12:59:54
358	58	7	1	150	2025-09-11 13:00:00	2025-09-11 13:00:00
359	58	1	1	144	2025-09-11 13:00:08	2025-09-11 13:00:08
360	58	3	1	224	2025-09-11 13:00:17	2025-09-11 13:00:17
361	60	10	1	170	2025-11-28 15:29:44	2025-11-28 15:29:44
362	60	8	1	128	2025-11-28 15:29:54	2025-11-28 15:29:54
363	60	9	1	71	2025-11-28 15:30:07	2025-11-28 15:30:07
364	60	11	1	98	2025-11-28 15:30:19	2025-11-28 15:30:19
365	60	1	1	107	2025-11-28 15:30:29	2025-11-28 15:30:29
366	61	10	1	179	2026-01-23 15:38:18	2026-01-23 15:38:18
367	61	8	1	155	2026-01-23 15:38:25	2026-01-23 15:38:25
368	61	4	1	78	2026-01-23 15:38:32	2026-01-23 15:38:32
369	61	9	1	72	2026-01-23 15:38:39	2026-01-23 15:38:39
370	61	2	1	124	2026-01-23 15:38:49	2026-01-23 15:38:49
371	61	7	1	203	2026-01-23 15:38:57	2026-01-23 15:38:57
372	61	1	1	121	2026-01-23 15:39:04	2026-01-23 15:39:04
373	61	3	1	197	2026-01-23 15:39:14	2026-01-23 15:39:14
\.


--
-- Data for Name: competition_types; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.competition_types (id, club_id, name, type, is_sex_specific, "position", description, created_at, updated_at) FROM stdin;
1	1	Fredenbaum	1	t	1	1. Fredenbaum des Abends	2025-03-08 11:38:54	2025-03-15 16:55:48
3	2	Fredenbaum	1	f	1	1. Fredenbaum	2025-03-24 16:43:47	2025-03-25 12:34:15
4	2	Fredenbaum 2	1	f	2	2. Fredenbaum	2025-03-24 16:44:16	2025-03-25 12:34:21
\.


--
-- Data for Name: dashboard_layouts; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.dashboard_layouts (id, user_id, club_id, layout, created_at, updated_at) FROM stdin;
7	1	1	[{"x":0,"y":0,"w":1,"h":1,"component":"Balance","props":{"club_id":1},"id":"0195ec23-460b-7317-a9d3-179cf355d459"},{"x":0,"y":1,"w":1,"h":1,"component":"Balance","props":{"player_id":1},"id":"0195ec23-61c2-744e-aea8-c18cd4dd9867"},{"x":1,"y":0,"w":2,"h":2,"component":"BalanceGraph","props":{"club_id":1},"id":"0195ec23-7601-701f-861e-bf304e7fab44"},{"x":1,"y":2,"w":2,"h":2,"component":"LastCompetition","props":{"competition_type_id":1},"id":"0195ec23-8e3c-7448-a185-cf18d20945d5"}]	2025-03-31 12:18:09	2025-05-28 13:36:43
8	1	2	[{"x":0,"y":0,"w":1,"h":1,"component":"Balance","props":{"club_id":2},"id":"0195ec24-31b2-70eb-a104-cb3ca648b4da"},{"x":1,"y":0,"w":2,"h":2,"component":"BalanceGraph","props":{"club_id":2},"id":"0195ec24-456a-708d-bc07-1970b98d7e57"},{"x":0,"y":1,"w":1,"h":1,"component":"Balance","props":{"player_id":35},"id":"0195ec24-6122-747b-a807-c78c8ecb7661"},{"x":1,"y":2,"w":2,"h":2,"component":"LastCompetition","props":{"competition_type_id":3},"id":"0195ec24-7ca3-7085-9ffe-7030e2c55fa9"}]	2025-03-31 12:19:09	2025-03-31 12:19:31
\.


--
-- Data for Name: failed_jobs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.failed_jobs (id, uuid, connection, queue, payload, exception, failed_at) FROM stdin;
\.


--
-- Data for Name: fee_entries; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.fee_entries (id, matchday_id, player_id, fee_type_version_id, amount, created_at, updated_at) FROM stdin;
529	17	10	3	0	2025-03-17 14:12:47	2025-03-17 14:12:47
530	17	10	5	0	2025-03-17 14:12:47	2025-03-17 14:12:47
531	17	10	4	0	2025-03-17 14:12:47	2025-03-17 14:12:47
532	17	10	6	0	2025-03-17 14:12:48	2025-03-17 14:12:48
534	17	10	19	0	2025-03-17 14:12:48	2025-03-17 14:12:48
536	17	10	8	0	2025-03-17 14:12:49	2025-03-17 14:12:49
538	17	8	5	0	2025-03-17 14:12:55	2025-03-17 14:12:55
539	17	8	4	0	2025-03-17 14:12:55	2025-03-17 14:12:55
540	17	8	6	0	2025-03-17 14:12:55	2025-03-17 14:12:55
542	17	8	19	0	2025-03-17 14:12:55	2025-03-17 14:12:55
544	17	8	8	0	2025-03-17 14:12:55	2025-03-17 14:12:55
545	17	12	3	0	2025-03-17 14:13:02	2025-03-17 14:13:02
546	17	12	5	0	2025-03-17 14:13:02	2025-03-17 14:13:02
547	17	12	4	0	2025-03-17 14:13:02	2025-03-17 14:13:02
548	17	12	6	0	2025-03-17 14:13:03	2025-03-17 14:13:03
550	17	12	19	0	2025-03-17 14:13:03	2025-03-17 14:13:03
552	17	12	8	0	2025-03-17 14:13:04	2025-03-17 14:13:04
553	17	4	3	0	2025-03-17 14:13:12	2025-03-17 14:13:12
554	17	4	5	0	2025-03-17 14:13:12	2025-03-17 14:13:12
555	17	4	4	0	2025-03-17 14:13:12	2025-03-17 14:13:12
558	17	4	19	0	2025-03-17 14:13:12	2025-03-17 14:13:12
560	17	4	8	0	2025-03-17 14:13:12	2025-03-17 14:13:12
561	17	9	3	0	2025-03-17 14:13:22	2025-03-17 14:13:22
562	17	9	5	0	2025-03-17 14:13:23	2025-03-17 14:13:23
563	17	9	4	0	2025-03-17 14:13:23	2025-03-17 14:13:23
564	17	9	6	0	2025-03-17 14:13:24	2025-03-17 14:13:24
566	17	9	19	0	2025-03-17 14:13:25	2025-03-17 14:13:25
568	17	9	8	0	2025-03-17 14:13:25	2025-03-17 14:13:25
569	17	13	3	0	2025-03-17 14:13:33	2025-03-17 14:13:33
570	17	13	5	0	2025-03-17 14:13:34	2025-03-17 14:13:34
571	17	13	4	0	2025-03-17 14:13:34	2025-03-17 14:13:34
572	17	13	6	0	2025-03-17 14:13:34	2025-03-17 14:13:34
574	17	13	19	0	2025-03-17 14:13:35	2025-03-17 14:13:35
576	17	13	8	0	2025-03-17 14:13:35	2025-03-17 14:13:35
577	17	2	3	0	2025-03-17 14:13:45	2025-03-17 14:13:45
578	17	2	5	0	2025-03-17 14:13:45	2025-03-17 14:13:45
579	17	2	4	0	2025-03-17 14:13:45	2025-03-17 14:13:45
580	17	2	6	0	2025-03-17 14:13:45	2025-03-17 14:13:45
582	17	2	19	0	2025-03-17 14:13:46	2025-03-17 14:13:46
584	17	2	8	0	2025-03-17 14:13:46	2025-03-17 14:13:46
585	17	7	3	0	2025-03-17 14:13:58	2025-03-17 14:13:58
586	17	7	5	0	2025-03-17 14:13:58	2025-03-17 14:13:58
587	17	7	4	0	2025-03-17 14:13:58	2025-03-17 14:13:58
588	17	7	6	0	2025-03-17 14:13:58	2025-03-17 14:13:58
590	17	7	19	0	2025-03-17 14:13:58	2025-03-17 14:13:58
592	17	7	8	0	2025-03-17 14:13:59	2025-03-17 14:13:59
593	17	1	3	0	2025-03-17 14:14:08	2025-03-17 14:14:08
594	17	1	5	0	2025-03-17 14:14:08	2025-03-17 14:14:08
595	17	1	4	0	2025-03-17 14:14:09	2025-03-17 14:14:09
596	17	1	6	0	2025-03-17 14:14:09	2025-03-17 14:14:09
598	17	1	19	0	2025-03-17 14:14:09	2025-03-17 14:14:09
600	17	1	8	0	2025-03-17 14:14:10	2025-03-17 14:14:10
535	17	10	1	4	2025-03-17 14:12:48	2025-03-17 14:14:51
533	17	10	2	5	2025-03-17 14:12:48	2025-03-17 14:14:51
543	17	8	1	7	2025-03-17 14:12:55	2025-03-17 14:15:21
541	17	8	2	4	2025-03-17 14:12:55	2025-03-17 14:15:22
537	17	8	3	1	2025-03-17 14:12:55	2025-03-17 14:15:23
551	17	12	1	2	2025-03-17 14:13:04	2025-03-17 14:15:42
549	17	12	2	6	2025-03-17 14:13:03	2025-03-17 14:15:43
559	17	4	1	11	2025-03-17 14:13:12	2025-03-17 14:16:01
557	17	4	2	5	2025-03-17 14:13:12	2025-03-17 14:16:02
567	17	9	1	6	2025-03-17 14:13:25	2025-03-17 14:16:26
565	17	9	2	3	2025-03-17 14:13:24	2025-03-17 14:16:27
556	17	4	6	1	2025-03-17 14:13:12	2025-03-17 14:16:36
575	17	13	1	7	2025-03-17 14:13:35	2025-03-17 14:17:04
573	17	13	2	5	2025-03-17 14:13:35	2025-03-17 14:17:05
583	17	2	1	7	2025-03-17 14:13:46	2025-03-17 14:17:21
581	17	2	2	7	2025-03-17 14:13:46	2025-03-17 14:17:21
591	17	7	1	10	2025-03-17 14:13:59	2025-03-17 14:17:41
589	17	7	2	9	2025-03-17 14:13:58	2025-03-17 14:17:41
599	17	1	1	4	2025-03-17 14:14:10	2025-03-17 14:18:00
597	17	1	2	8	2025-03-17 14:14:09	2025-03-17 14:18:01
673	19	10	3	0	2025-03-18 07:09:28	2025-03-18 07:09:28
674	19	10	5	0	2025-03-18 07:09:28	2025-03-18 07:09:28
675	19	10	4	0	2025-03-18 07:09:28	2025-03-18 07:09:28
676	19	10	6	0	2025-03-18 07:09:28	2025-03-18 07:09:28
678	19	10	19	0	2025-03-18 07:09:28	2025-03-18 07:09:28
969	24	10	3	0	2025-03-18 20:24:55	2025-03-18 20:24:55
680	19	10	8	0	2025-03-18 07:09:28	2025-03-18 07:09:28
681	19	7	3	0	2025-03-18 07:09:33	2025-03-18 07:09:33
677	19	10	2	7	2025-03-18 07:09:28	2025-03-18 07:10:16
970	24	10	5	0	2025-03-18 20:24:55	2025-03-18 20:24:55
971	24	10	4	0	2025-03-18 20:24:56	2025-03-18 20:24:56
972	24	10	6	0	2025-03-18 20:24:56	2025-03-18 20:24:56
974	24	10	19	0	2025-03-18 20:24:57	2025-03-18 20:24:57
976	24	10	8	0	2025-03-18 20:24:57	2025-03-18 20:24:57
979	24	9	3	0	2025-03-18 20:25:01	2025-03-18 20:25:01
980	24	9	4	0	2025-03-18 20:25:01	2025-03-18 20:25:01
981	24	9	5	0	2025-03-18 20:25:01	2025-03-18 20:25:01
982	24	9	6	0	2025-03-18 20:25:01	2025-03-18 20:25:01
983	24	9	8	0	2025-03-18 20:25:01	2025-03-18 20:25:01
984	24	9	19	0	2025-03-18 20:25:01	2025-03-18 20:25:01
973	24	10	2	9	2025-03-18 20:24:57	2025-03-18 20:25:46
977	24	9	1	4	2025-03-18 20:25:01	2025-03-18 20:25:55
978	24	9	2	6	2025-03-18 20:25:01	2025-03-18 20:25:56
975	24	10	1	6	2025-03-18 20:24:57	2025-03-18 20:25:45
601	17	3	3	0	2025-03-17 14:14:19	2025-03-17 14:14:19
602	17	3	5	0	2025-03-17 14:14:20	2025-03-17 14:14:20
603	17	3	4	0	2025-03-17 14:14:20	2025-03-17 14:14:20
604	17	3	6	0	2025-03-17 14:14:20	2025-03-17 14:14:20
606	17	3	19	0	2025-03-17 14:14:21	2025-03-17 14:14:21
608	17	3	8	0	2025-03-17 14:14:21	2025-03-17 14:14:21
607	17	3	1	3	2025-03-17 14:14:21	2025-03-17 14:18:15
605	17	3	2	3	2025-03-17 14:14:21	2025-03-17 14:18:16
609	18	3	3	0	2025-03-18 07:05:39	2025-03-18 07:05:39
610	18	3	5	0	2025-03-18 07:05:39	2025-03-18 07:05:39
611	18	3	4	0	2025-03-18 07:05:39	2025-03-18 07:05:39
612	18	3	6	0	2025-03-18 07:05:39	2025-03-18 07:05:39
614	18	3	19	0	2025-03-18 07:05:39	2025-03-18 07:05:39
616	18	3	8	0	2025-03-18 07:05:39	2025-03-18 07:05:39
615	18	3	1	1	2025-03-18 07:05:39	2025-03-18 07:05:46
613	18	3	2	2	2025-03-18 07:05:39	2025-03-18 07:05:47
620	18	8	4	0	2025-03-18 07:06:01	2025-03-18 07:06:01
621	18	8	5	0	2025-03-18 07:06:01	2025-03-18 07:06:01
622	18	8	6	0	2025-03-18 07:06:01	2025-03-18 07:06:01
623	18	8	8	0	2025-03-18 07:06:01	2025-03-18 07:06:01
624	18	8	19	0	2025-03-18 07:06:01	2025-03-18 07:06:01
617	18	8	1	7	2025-03-18 07:06:00	2025-03-18 07:06:12
618	18	8	2	6	2025-03-18 07:06:01	2025-03-18 07:06:12
619	18	8	3	2	2025-03-18 07:06:01	2025-03-18 07:06:13
627	18	4	3	0	2025-03-18 07:06:36	2025-03-18 07:06:36
628	18	4	4	0	2025-03-18 07:06:36	2025-03-18 07:06:36
629	18	4	5	0	2025-03-18 07:06:36	2025-03-18 07:06:36
630	18	4	6	0	2025-03-18 07:06:36	2025-03-18 07:06:36
631	18	4	8	0	2025-03-18 07:06:36	2025-03-18 07:06:36
632	18	4	19	0	2025-03-18 07:06:36	2025-03-18 07:06:36
635	18	9	3	0	2025-03-18 07:06:39	2025-03-18 07:06:39
636	18	9	4	0	2025-03-18 07:06:39	2025-03-18 07:06:39
637	18	9	5	0	2025-03-18 07:06:39	2025-03-18 07:06:39
638	18	9	6	0	2025-03-18 07:06:39	2025-03-18 07:06:39
639	18	9	8	0	2025-03-18 07:06:39	2025-03-18 07:06:39
640	18	9	19	0	2025-03-18 07:06:39	2025-03-18 07:06:39
626	18	4	2	3	2025-03-18 07:06:34	2025-03-18 07:07:16
633	18	9	1	6	2025-03-18 07:06:39	2025-03-18 07:07:31
634	18	9	2	4	2025-03-18 07:06:39	2025-03-18 07:07:31
641	18	11	1	3	2025-03-18 07:06:44	2025-03-18 07:08:08
1464	35	30	31	0	2025-03-27 15:25:39	2025-03-27 15:25:39
1465	35	30	32	0	2025-03-27 15:25:39	2025-03-27 15:25:39
1468	35	30	35	0	2025-03-27 15:25:40	2025-03-27 15:25:40
1471	35	36	31	0	2025-03-27 15:26:07	2025-03-27 15:26:07
1473	35	36	33	0	2025-03-27 15:26:07	2025-03-27 15:26:07
1474	35	36	34	0	2025-03-27 15:26:07	2025-03-27 15:26:07
1475	35	36	35	0	2025-03-27 15:26:07	2025-03-27 15:26:07
1478	35	32	31	0	2025-03-27 15:26:10	2025-03-27 15:26:10
1480	35	32	33	0	2025-03-27 15:26:10	2025-03-27 15:26:10
1481	35	32	34	0	2025-03-27 15:26:10	2025-03-27 15:26:10
1482	35	32	35	0	2025-03-27 15:26:10	2025-03-27 15:26:10
1485	35	34	31	0	2025-03-27 15:26:14	2025-03-27 15:26:14
1486	35	34	32	0	2025-03-27 15:26:14	2025-03-27 15:26:14
1487	35	34	33	0	2025-03-27 15:26:15	2025-03-27 15:26:15
1489	35	34	34	0	2025-03-27 15:26:15	2025-03-27 15:26:15
1490	35	34	35	0	2025-03-27 15:26:16	2025-03-27 15:26:16
1506	35	37	31	0	2025-03-27 15:26:20	2025-03-27 15:26:20
1508	35	37	33	0	2025-03-27 15:26:20	2025-03-27 15:26:20
1510	35	37	34	0	2025-03-27 15:26:21	2025-03-27 15:26:21
1512	35	37	35	0	2025-03-27 15:26:21	2025-03-27 15:26:21
1520	36	36	31	0	2025-03-27 15:27:44	2025-03-27 15:27:44
1522	36	36	33	0	2025-03-27 15:27:45	2025-03-27 15:27:45
1523	36	36	34	0	2025-03-27 15:27:45	2025-03-27 15:27:45
1524	36	36	35	0	2025-03-27 15:27:45	2025-03-27 15:27:45
1527	36	37	31	0	2025-03-27 15:27:53	2025-03-27 15:27:53
1528	36	37	32	0	2025-03-27 15:27:53	2025-03-27 15:27:53
1529	36	37	33	0	2025-03-27 15:27:53	2025-03-27 15:27:53
1530	36	37	34	0	2025-03-27 15:27:53	2025-03-27 15:27:53
1531	36	37	35	0	2025-03-27 15:27:53	2025-03-27 15:27:53
1534	36	31	31	0	2025-03-27 15:27:57	2025-03-27 15:27:57
1535	36	31	32	0	2025-03-27 15:27:57	2025-03-27 15:27:57
1536	36	31	33	0	2025-03-27 15:27:57	2025-03-27 15:27:57
1537	36	31	34	0	2025-03-27 15:27:57	2025-03-27 15:27:57
1538	36	31	35	0	2025-03-27 15:27:58	2025-03-27 15:27:58
1463	35	30	30	9	2025-03-27 15:25:39	2025-03-27 15:37:03
1467	35	30	34	1	2025-03-27 15:25:40	2025-03-27 15:37:04
1469	35	36	29	3	2025-03-27 15:26:07	2025-03-27 15:38:06
1470	35	36	30	5	2025-03-27 15:26:07	2025-03-27 15:38:06
1472	35	36	32	1	2025-03-27 15:26:07	2025-03-27 15:38:07
1477	35	32	30	8	2025-03-27 15:26:09	2025-03-27 15:38:21
1479	35	32	32	1	2025-03-27 15:26:10	2025-03-27 15:38:21
1483	35	34	29	5	2025-03-27 15:26:13	2025-03-27 15:39:12
1484	35	34	30	5	2025-03-27 15:26:14	2025-03-27 15:39:13
1504	35	37	29	3	2025-03-27 15:26:19	2025-03-27 15:40:05
1505	35	37	30	3	2025-03-27 15:26:19	2025-03-27 15:40:05
1507	35	37	32	1	2025-03-27 15:26:20	2025-03-27 15:40:06
1466	35	30	33	1	2025-03-27 15:25:40	2025-04-25 12:13:07
1532	36	31	29	10	2025-03-27 15:27:57	2025-04-25 12:14:41
1533	36	31	30	10	2025-03-27 15:27:57	2025-04-25 12:14:41
1539	36	33	29	8	2025-03-27 15:28:02	2025-04-25 12:14:55
1540	36	33	30	9	2025-03-27 15:28:02	2025-04-25 12:14:55
1518	36	36	29	5	2025-03-27 15:27:44	2025-04-25 12:15:10
1519	36	36	30	6	2025-03-27 15:27:44	2025-04-25 12:15:10
1521	36	36	32	2	2025-03-27 15:27:45	2025-04-25 12:15:10
1525	36	37	29	3	2025-03-27 15:27:53	2025-04-25 12:15:24
1526	36	37	30	5	2025-03-27 15:27:53	2025-04-25 12:15:24
1492	35	35	31	0	2025-03-27 15:26:16	2025-03-27 15:26:16
1493	35	35	32	0	2025-03-27 15:26:16	2025-03-27 15:26:16
1495	35	35	33	0	2025-03-27 15:26:17	2025-03-27 15:26:17
1497	35	35	34	0	2025-03-27 15:26:17	2025-03-27 15:26:17
1499	35	35	35	0	2025-03-27 15:26:17	2025-03-27 15:26:17
1513	35	31	31	0	2025-03-27 15:26:21	2025-03-27 15:26:21
1514	35	31	32	0	2025-03-27 15:26:22	2025-03-27 15:26:22
1515	35	31	33	0	2025-03-27 15:26:22	2025-03-27 15:26:22
1516	35	31	34	0	2025-03-27 15:26:23	2025-03-27 15:26:23
1517	35	31	35	0	2025-03-27 15:26:23	2025-03-27 15:26:23
1462	35	30	29	7	2025-03-27 15:25:39	2025-03-27 15:37:03
1476	35	32	29	8	2025-03-27 15:26:09	2025-03-27 15:38:20
1488	35	35	29	7	2025-03-27 15:26:15	2025-03-27 15:39:45
1491	35	35	30	7	2025-03-27 15:26:16	2025-03-27 15:39:45
1509	35	31	29	8	2025-03-27 15:26:21	2025-03-27 15:40:36
1511	35	31	30	9	2025-03-27 15:26:21	2025-03-27 15:40:36
1546	45	36	29	2	2025-04-25 12:06:12	2025-04-25 12:06:12
1547	45	36	30	2	2025-04-25 12:06:13	2025-04-25 12:06:13
481	16	7	3	0	2025-03-17 14:00:02	2025-03-17 14:00:02
482	16	7	5	0	2025-03-17 14:00:03	2025-03-17 14:00:03
483	16	7	4	0	2025-03-17 14:00:03	2025-03-17 14:00:03
484	16	7	6	0	2025-03-17 14:00:03	2025-03-17 14:00:03
486	16	7	19	0	2025-03-17 14:00:05	2025-03-17 14:00:05
488	16	7	8	0	2025-03-17 14:00:06	2025-03-17 14:00:06
489	16	8	3	0	2025-03-17 14:00:15	2025-03-17 14:00:15
490	16	8	5	0	2025-03-17 14:00:15	2025-03-17 14:00:15
491	16	8	4	0	2025-03-17 14:00:15	2025-03-17 14:00:15
492	16	8	6	0	2025-03-17 14:00:15	2025-03-17 14:00:15
494	16	8	19	0	2025-03-17 14:00:15	2025-03-17 14:00:15
496	16	8	8	0	2025-03-17 14:00:15	2025-03-17 14:00:15
497	16	4	3	0	2025-03-17 14:00:25	2025-03-17 14:00:25
498	16	4	5	0	2025-03-17 14:00:25	2025-03-17 14:00:25
499	16	4	4	0	2025-03-17 14:00:25	2025-03-17 14:00:25
500	16	4	6	0	2025-03-17 14:00:25	2025-03-17 14:00:25
502	16	4	19	0	2025-03-17 14:00:25	2025-03-17 14:00:25
504	16	4	8	0	2025-03-17 14:00:25	2025-03-17 14:00:25
485	16	7	2	8	2025-03-17 14:00:04	2025-03-17 14:01:46
495	16	8	1	3	2025-03-17 14:00:15	2025-03-17 14:02:24
493	16	8	2	6	2025-03-17 14:00:15	2025-03-17 14:02:24
503	16	4	1	9	2025-03-17 14:00:25	2025-03-17 14:02:57
501	16	4	2	6	2025-03-17 14:00:25	2025-03-17 14:02:57
505	16	9	3	0	2025-03-17 14:00:34	2025-03-17 14:00:34
507	16	9	4	0	2025-03-17 14:00:34	2025-03-17 14:00:34
508	16	9	6	0	2025-03-17 14:00:35	2025-03-17 14:00:35
510	16	9	19	0	2025-03-17 14:00:35	2025-03-17 14:00:35
512	16	9	8	0	2025-03-17 14:00:35	2025-03-17 14:00:35
513	16	2	3	0	2025-03-17 14:00:44	2025-03-17 14:00:44
514	16	2	5	0	2025-03-17 14:00:45	2025-03-17 14:00:45
515	16	2	4	0	2025-03-17 14:00:45	2025-03-17 14:00:45
516	16	2	6	0	2025-03-17 14:00:46	2025-03-17 14:00:46
518	16	2	19	0	2025-03-17 14:00:46	2025-03-17 14:00:46
520	16	2	8	0	2025-03-17 14:00:47	2025-03-17 14:00:47
521	16	3	3	0	2025-03-17 14:00:55	2025-03-17 14:00:55
522	16	3	5	0	2025-03-17 14:00:55	2025-03-17 14:00:55
523	16	3	4	0	2025-03-17 14:00:55	2025-03-17 14:00:55
524	16	3	6	0	2025-03-17 14:00:55	2025-03-17 14:00:55
526	16	3	19	0	2025-03-17 14:00:56	2025-03-17 14:00:56
528	16	3	8	0	2025-03-17 14:00:56	2025-03-17 14:00:56
487	16	7	1	7	2025-03-17 14:00:05	2025-03-17 14:01:46
511	16	9	1	9	2025-03-17 14:00:35	2025-03-17 14:03:20
509	16	9	2	13	2025-03-17 14:00:35	2025-03-17 14:03:22
519	16	2	1	4	2025-03-17 14:00:47	2025-03-17 14:03:40
517	16	2	2	7	2025-03-17 14:00:46	2025-03-17 14:03:41
527	16	3	1	4	2025-03-17 14:00:56	2025-03-17 14:04:02
525	16	3	2	4	2025-03-17 14:00:55	2025-03-17 14:04:03
506	16	9	5	1	2025-03-17 14:00:34	2025-03-17 14:05:40
643	18	11	3	0	2025-03-18 07:06:45	2025-03-18 07:06:45
644	18	11	4	0	2025-03-18 07:06:45	2025-03-18 07:06:45
645	18	11	5	0	2025-03-18 07:06:45	2025-03-18 07:06:45
646	18	11	6	0	2025-03-18 07:06:45	2025-03-18 07:06:45
647	18	11	8	0	2025-03-18 07:06:46	2025-03-18 07:06:46
648	18	11	19	0	2025-03-18 07:06:46	2025-03-18 07:06:46
651	18	2	3	0	2025-03-18 07:06:51	2025-03-18 07:06:51
652	18	2	4	0	2025-03-18 07:06:51	2025-03-18 07:06:51
653	18	2	5	0	2025-03-18 07:06:51	2025-03-18 07:06:51
654	18	2	6	0	2025-03-18 07:06:51	2025-03-18 07:06:51
655	18	2	8	0	2025-03-18 07:06:51	2025-03-18 07:06:51
656	18	2	19	0	2025-03-18 07:06:52	2025-03-18 07:06:52
659	18	1	3	0	2025-03-18 07:06:55	2025-03-18 07:06:55
660	18	1	4	0	2025-03-18 07:06:55	2025-03-18 07:06:55
661	18	1	5	0	2025-03-18 07:06:55	2025-03-18 07:06:55
662	18	1	6	0	2025-03-18 07:06:55	2025-03-18 07:06:55
663	18	1	8	0	2025-03-18 07:06:55	2025-03-18 07:06:55
664	18	1	19	0	2025-03-18 07:06:55	2025-03-18 07:06:55
667	18	10	3	0	2025-03-18 07:07:05	2025-03-18 07:07:05
668	18	10	4	0	2025-03-18 07:07:05	2025-03-18 07:07:05
669	18	10	5	0	2025-03-18 07:07:05	2025-03-18 07:07:05
670	18	10	6	0	2025-03-18 07:07:06	2025-03-18 07:07:06
671	18	10	8	0	2025-03-18 07:07:06	2025-03-18 07:07:06
672	18	10	19	0	2025-03-18 07:07:07	2025-03-18 07:07:07
625	18	4	1	6	2025-03-18 07:06:34	2025-03-18 07:07:16
665	18	10	1	5	2025-03-18 07:07:05	2025-03-18 07:07:22
666	18	10	2	6	2025-03-18 07:07:05	2025-03-18 07:07:23
1548	45	36	31	0	2025-04-25 12:06:13	2025-04-25 12:06:13
1549	45	36	32	1	2025-04-25 12:06:13	2025-04-25 12:06:13
1550	45	36	33	0	2025-04-25 12:06:14	2025-04-25 12:06:14
1551	45	36	34	0	2025-04-25 12:06:14	2025-04-25 12:06:14
649	18	2	1	3	2025-03-18 07:06:50	2025-03-18 07:07:40
650	18	2	2	5	2025-03-18 07:06:50	2025-03-18 07:07:41
657	18	1	1	1	2025-03-18 07:06:55	2025-03-18 07:07:57
658	18	1	2	7	2025-03-18 07:06:55	2025-03-18 07:07:58
642	18	11	2	9	2025-03-18 07:06:44	2025-03-18 07:08:09
683	19	7	4	0	2025-03-18 07:09:34	2025-03-18 07:09:34
684	19	7	6	0	2025-03-18 07:09:35	2025-03-18 07:09:35
686	19	7	19	0	2025-03-18 07:09:35	2025-03-18 07:09:35
688	19	7	8	0	2025-03-18 07:09:36	2025-03-18 07:09:36
690	19	11	5	0	2025-03-18 07:09:40	2025-03-18 07:09:40
691	19	11	4	0	2025-03-18 07:09:41	2025-03-18 07:09:41
692	19	11	6	0	2025-03-18 07:09:41	2025-03-18 07:09:41
694	19	11	19	0	2025-03-18 07:09:41	2025-03-18 07:09:41
696	19	11	8	0	2025-03-18 07:09:41	2025-03-18 07:09:41
697	19	2	3	0	2025-03-18 07:09:44	2025-03-18 07:09:44
699	19	2	4	0	2025-03-18 07:09:45	2025-03-18 07:09:45
700	19	2	6	0	2025-03-18 07:09:45	2025-03-18 07:09:45
702	19	2	19	0	2025-03-18 07:09:46	2025-03-18 07:09:46
704	19	2	8	0	2025-03-18 07:09:46	2025-03-18 07:09:46
706	19	1	5	0	2025-03-18 07:09:51	2025-03-18 07:09:51
707	19	1	4	0	2025-03-18 07:09:51	2025-03-18 07:09:51
708	19	1	6	0	2025-03-18 07:09:51	2025-03-18 07:09:51
710	19	1	19	0	2025-03-18 07:09:52	2025-03-18 07:09:52
712	19	1	8	0	2025-03-18 07:09:52	2025-03-18 07:09:52
687	19	7	1	6	2025-03-18 07:09:35	2025-03-18 07:10:06
685	19	7	2	6	2025-03-18 07:09:35	2025-03-18 07:10:07
682	19	7	5	2	2025-03-18 07:09:34	2025-03-18 07:10:08
679	19	10	1	6	2025-03-18 07:09:28	2025-03-18 07:10:15
703	19	2	1	8	2025-03-18 07:09:46	2025-03-18 07:10:31
701	19	2	2	4	2025-03-18 07:09:45	2025-03-18 07:10:31
698	19	2	5	2	2025-03-18 07:09:45	2025-03-18 07:10:31
711	19	1	1	8	2025-03-18 07:09:52	2025-03-18 07:10:40
709	19	1	2	7	2025-03-18 07:09:52	2025-03-18 07:10:41
705	19	1	3	1	2025-03-18 07:09:51	2025-03-18 07:10:42
695	19	11	1	14	2025-03-18 07:09:41	2025-03-18 07:10:51
693	19	11	2	15	2025-03-18 07:09:41	2025-03-18 07:10:52
689	19	11	3	2	2025-03-18 07:09:40	2025-03-18 07:10:52
713	20	10	3	0	2025-03-18 07:12:03	2025-03-18 07:12:03
714	20	10	5	0	2025-03-18 07:12:03	2025-03-18 07:12:03
715	20	10	4	0	2025-03-18 07:12:03	2025-03-18 07:12:03
716	20	10	6	0	2025-03-18 07:12:03	2025-03-18 07:12:03
718	20	10	19	0	2025-03-18 07:12:03	2025-03-18 07:12:03
720	20	10	8	0	2025-03-18 07:12:03	2025-03-18 07:12:03
722	20	8	5	0	2025-03-18 07:12:06	2025-03-18 07:12:06
723	20	8	4	0	2025-03-18 07:12:06	2025-03-18 07:12:06
724	20	8	6	0	2025-03-18 07:12:07	2025-03-18 07:12:07
726	20	8	19	0	2025-03-18 07:12:07	2025-03-18 07:12:07
728	20	8	8	0	2025-03-18 07:12:07	2025-03-18 07:12:07
730	20	4	5	0	2025-03-18 07:12:11	2025-03-18 07:12:11
731	20	4	4	0	2025-03-18 07:12:12	2025-03-18 07:12:12
732	20	4	6	0	2025-03-18 07:12:12	2025-03-18 07:12:12
734	20	4	19	0	2025-03-18 07:12:13	2025-03-18 07:12:13
736	20	4	8	0	2025-03-18 07:12:13	2025-03-18 07:12:13
738	20	11	5	0	2025-03-18 07:12:16	2025-03-18 07:12:16
739	20	11	4	0	2025-03-18 07:12:16	2025-03-18 07:12:16
740	20	11	6	0	2025-03-18 07:12:16	2025-03-18 07:12:16
742	20	11	19	0	2025-03-18 07:12:16	2025-03-18 07:12:16
744	20	11	8	0	2025-03-18 07:12:17	2025-03-18 07:12:17
745	20	2	3	0	2025-03-18 07:12:23	2025-03-18 07:12:23
746	20	2	5	0	2025-03-18 07:12:23	2025-03-18 07:12:23
747	20	2	4	0	2025-03-18 07:12:24	2025-03-18 07:12:24
748	20	2	6	0	2025-03-18 07:12:24	2025-03-18 07:12:24
750	20	2	19	0	2025-03-18 07:12:25	2025-03-18 07:12:25
752	20	2	8	0	2025-03-18 07:12:25	2025-03-18 07:12:25
753	20	3	3	0	2025-03-18 07:12:29	2025-03-18 07:12:29
754	20	3	5	0	2025-03-18 07:12:29	2025-03-18 07:12:29
755	20	3	4	0	2025-03-18 07:12:29	2025-03-18 07:12:29
756	20	3	6	0	2025-03-18 07:12:29	2025-03-18 07:12:29
758	20	3	19	0	2025-03-18 07:12:29	2025-03-18 07:12:29
760	20	3	8	0	2025-03-18 07:12:30	2025-03-18 07:12:30
735	20	4	1	8	2025-03-18 07:12:13	2025-03-18 07:12:47
733	20	4	2	8	2025-03-18 07:12:12	2025-03-18 07:12:48
729	20	4	3	2	2025-03-18 07:12:11	2025-03-18 07:12:48
793	21	7	3	0	2025-03-18 07:16:13	2025-03-18 07:16:13
727	20	8	1	3	2025-03-18 07:12:07	2025-03-18 07:13:04
725	20	8	2	2	2025-03-18 07:12:07	2025-03-18 07:13:05
721	20	8	3	1	2025-03-18 07:12:06	2025-03-18 07:13:06
719	20	10	1	4	2025-03-18 07:12:03	2025-03-18 07:13:14
717	20	10	2	5	2025-03-18 07:12:03	2025-03-18 07:13:15
751	20	2	1	7	2025-03-18 07:12:25	2025-03-18 07:13:34
749	20	2	2	6	2025-03-18 07:12:24	2025-03-18 07:13:36
743	20	11	1	7	2025-03-18 07:12:17	2025-03-18 07:13:45
741	20	11	2	9	2025-03-18 07:12:16	2025-03-18 07:13:46
737	20	11	3	1	2025-03-18 07:12:16	2025-03-18 07:13:47
759	20	3	1	2	2025-03-18 07:12:29	2025-03-18 07:13:58
757	20	3	2	2	2025-03-18 07:12:29	2025-03-18 07:13:58
761	21	10	3	0	2025-03-18 07:15:50	2025-03-18 07:15:50
762	21	10	5	0	2025-03-18 07:15:50	2025-03-18 07:15:50
763	21	10	4	0	2025-03-18 07:15:50	2025-03-18 07:15:50
764	21	10	6	0	2025-03-18 07:15:51	2025-03-18 07:15:51
765	21	10	2	0	2025-03-18 07:15:51	2025-03-18 07:15:51
768	21	10	8	0	2025-03-18 07:15:52	2025-03-18 07:15:52
770	21	11	5	0	2025-03-18 07:15:57	2025-03-18 07:15:57
771	21	11	4	0	2025-03-18 07:15:57	2025-03-18 07:15:57
772	21	11	6	0	2025-03-18 07:15:57	2025-03-18 07:15:57
776	21	11	8	0	2025-03-18 07:15:58	2025-03-18 07:15:58
777	21	2	3	0	2025-03-18 07:16:02	2025-03-18 07:16:02
778	21	2	5	0	2025-03-18 07:16:02	2025-03-18 07:16:02
779	21	2	4	0	2025-03-18 07:16:02	2025-03-18 07:16:02
780	21	2	6	0	2025-03-18 07:16:02	2025-03-18 07:16:02
782	21	2	19	0	2025-03-18 07:16:02	2025-03-18 07:16:02
784	21	2	8	0	2025-03-18 07:16:02	2025-03-18 07:16:02
785	21	1	3	0	2025-03-18 07:16:07	2025-03-18 07:16:07
786	21	1	5	0	2025-03-18 07:16:07	2025-03-18 07:16:07
787	21	1	4	0	2025-03-18 07:16:08	2025-03-18 07:16:08
788	21	1	6	0	2025-03-18 07:16:08	2025-03-18 07:16:08
792	21	1	8	0	2025-03-18 07:16:09	2025-03-18 07:16:09
794	21	7	5	0	2025-03-18 07:16:13	2025-03-18 07:16:13
795	21	7	4	0	2025-03-18 07:16:14	2025-03-18 07:16:14
796	21	7	6	0	2025-03-18 07:16:14	2025-03-18 07:16:14
798	21	7	19	0	2025-03-18 07:16:14	2025-03-18 07:16:14
800	21	7	8	0	2025-03-18 07:16:14	2025-03-18 07:16:14
799	21	7	1	5	2025-03-18 07:16:14	2025-03-18 07:16:22
797	21	7	2	6	2025-03-18 07:16:14	2025-03-18 07:16:22
767	21	10	1	3	2025-03-18 07:15:52	2025-03-18 07:16:39
766	21	10	19	1	2025-03-18 07:15:51	2025-03-18 07:16:39
783	21	2	1	5	2025-03-18 07:16:02	2025-03-18 07:16:49
781	21	2	2	1	2025-03-18 07:16:02	2025-03-18 07:16:49
791	21	1	1	9	2025-03-18 07:16:08	2025-03-18 07:17:00
789	21	1	2	8	2025-03-18 07:16:08	2025-03-18 07:17:01
790	21	1	19	1	2025-03-18 07:16:08	2025-03-18 07:17:01
775	21	11	1	8	2025-03-18 07:15:58	2025-03-18 07:17:13
773	21	11	2	12	2025-03-18 07:15:57	2025-03-18 07:17:14
769	21	11	3	1	2025-03-18 07:15:56	2025-03-18 07:17:15
774	21	11	19	1	2025-03-18 07:15:58	2025-03-18 07:17:15
801	22	10	3	0	2025-03-18 07:18:58	2025-03-18 07:18:58
802	22	10	5	0	2025-03-18 07:18:58	2025-03-18 07:18:58
803	22	10	4	0	2025-03-18 07:18:58	2025-03-18 07:18:58
804	22	10	6	0	2025-03-18 07:18:58	2025-03-18 07:18:58
806	22	10	19	0	2025-03-18 07:18:58	2025-03-18 07:18:58
808	22	10	8	0	2025-03-18 07:18:58	2025-03-18 07:18:58
809	22	9	3	0	2025-03-18 07:19:01	2025-03-18 07:19:01
810	22	9	5	0	2025-03-18 07:19:02	2025-03-18 07:19:02
811	22	9	4	0	2025-03-18 07:19:02	2025-03-18 07:19:02
812	22	9	6	0	2025-03-18 07:19:02	2025-03-18 07:19:02
813	22	9	2	6	2025-03-18 07:19:02	2025-03-18 07:20:29
816	22	9	8	0	2025-03-18 07:19:03	2025-03-18 07:19:03
817	22	14	3	0	2025-03-18 07:19:07	2025-03-18 07:19:07
818	22	14	5	0	2025-03-18 07:19:07	2025-03-18 07:19:07
819	22	14	4	0	2025-03-18 07:19:07	2025-03-18 07:19:07
820	22	14	6	0	2025-03-18 07:19:08	2025-03-18 07:19:08
822	22	14	19	0	2025-03-18 07:19:08	2025-03-18 07:19:08
824	22	14	8	0	2025-03-18 07:19:09	2025-03-18 07:19:09
825	22	2	3	0	2025-03-18 07:19:12	2025-03-18 07:19:12
826	22	2	5	0	2025-03-18 07:19:12	2025-03-18 07:19:12
827	22	2	4	0	2025-03-18 07:19:12	2025-03-18 07:19:12
828	22	2	6	0	2025-03-18 07:19:12	2025-03-18 07:19:12
830	22	2	19	0	2025-03-18 07:19:12	2025-03-18 07:19:12
832	22	2	8	0	2025-03-18 07:19:12	2025-03-18 07:19:12
834	22	1	5	0	2025-03-18 07:19:17	2025-03-18 07:19:17
835	22	1	4	0	2025-03-18 07:19:17	2025-03-18 07:19:17
836	22	1	6	0	2025-03-18 07:19:17	2025-03-18 07:19:17
838	22	1	19	0	2025-03-18 07:19:18	2025-03-18 07:19:18
840	22	1	8	0	2025-03-18 07:19:18	2025-03-18 07:19:18
841	22	5	3	0	2025-03-18 07:19:22	2025-03-18 07:19:22
842	22	5	5	0	2025-03-18 07:19:22	2025-03-18 07:19:22
843	22	5	4	0	2025-03-18 07:19:22	2025-03-18 07:19:22
844	22	5	6	0	2025-03-18 07:19:22	2025-03-18 07:19:22
845	22	5	2	0	2025-03-18 07:19:22	2025-03-18 07:19:22
846	22	5	19	0	2025-03-18 07:19:23	2025-03-18 07:19:23
847	22	5	1	0	2025-03-18 07:19:23	2025-03-18 07:19:23
848	22	5	8	0	2025-03-18 07:19:23	2025-03-18 07:19:23
823	22	14	1	7	2025-03-18 07:19:08	2025-03-18 07:19:40
821	22	14	2	3	2025-03-18 07:19:08	2025-03-18 07:19:41
807	22	10	1	2	2025-03-18 07:18:58	2025-03-18 07:20:16
805	22	10	2	5	2025-03-18 07:18:58	2025-03-18 07:20:17
815	22	9	1	7	2025-03-18 07:19:03	2025-03-18 07:20:29
814	22	9	19	1	2025-03-18 07:19:02	2025-03-18 07:20:29
831	22	2	1	6	2025-03-18 07:19:12	2025-03-18 07:20:42
829	22	2	2	5	2025-03-18 07:19:12	2025-03-18 07:20:43
839	22	1	1	5	2025-03-18 07:19:18	2025-03-18 07:20:51
837	22	1	2	8	2025-03-18 07:19:17	2025-03-18 07:20:51
833	22	1	3	1	2025-03-18 07:19:17	2025-03-18 07:20:51
849	23	4	3	0	2025-03-18 20:18:46	2025-03-18 20:18:46
850	23	4	5	0	2025-03-18 20:18:46	2025-03-18 20:18:46
851	23	4	4	0	2025-03-18 20:18:46	2025-03-18 20:18:46
852	23	4	6	0	2025-03-18 20:18:46	2025-03-18 20:18:46
854	23	4	19	0	2025-03-18 20:18:46	2025-03-18 20:18:46
856	23	4	8	0	2025-03-18 20:18:46	2025-03-18 20:18:46
857	23	8	3	0	2025-03-18 20:18:51	2025-03-18 20:18:51
858	23	8	5	0	2025-03-18 20:18:51	2025-03-18 20:18:51
859	23	8	4	0	2025-03-18 20:18:51	2025-03-18 20:18:51
860	23	8	6	0	2025-03-18 20:18:51	2025-03-18 20:18:51
862	23	8	19	0	2025-03-18 20:18:52	2025-03-18 20:18:52
864	23	8	8	0	2025-03-18 20:18:52	2025-03-18 20:18:52
865	23	3	3	0	2025-03-18 20:19:00	2025-03-18 20:19:00
866	23	3	5	0	2025-03-18 20:19:00	2025-03-18 20:19:00
867	23	3	4	0	2025-03-18 20:19:01	2025-03-18 20:19:01
868	23	3	6	0	2025-03-18 20:19:01	2025-03-18 20:19:01
870	23	3	19	0	2025-03-18 20:19:02	2025-03-18 20:19:02
872	23	3	8	0	2025-03-18 20:19:02	2025-03-18 20:19:02
874	23	11	5	0	2025-03-18 20:19:07	2025-03-18 20:19:07
875	23	11	4	0	2025-03-18 20:19:07	2025-03-18 20:19:07
876	23	11	6	0	2025-03-18 20:19:07	2025-03-18 20:19:07
878	23	11	19	0	2025-03-18 20:19:08	2025-03-18 20:19:08
880	23	11	8	0	2025-03-18 20:19:08	2025-03-18 20:19:08
881	23	10	3	0	2025-03-18 20:19:11	2025-03-18 20:19:11
882	23	10	5	0	2025-03-18 20:19:12	2025-03-18 20:19:12
883	23	10	4	0	2025-03-18 20:19:12	2025-03-18 20:19:12
884	23	10	6	0	2025-03-18 20:19:12	2025-03-18 20:19:12
886	23	10	19	0	2025-03-18 20:19:12	2025-03-18 20:19:12
888	23	10	8	0	2025-03-18 20:19:12	2025-03-18 20:19:12
889	23	13	3	0	2025-03-18 20:19:18	2025-03-18 20:19:18
890	23	13	5	0	2025-03-18 20:19:18	2025-03-18 20:19:18
891	23	13	4	0	2025-03-18 20:19:19	2025-03-18 20:19:19
892	23	13	6	0	2025-03-18 20:19:19	2025-03-18 20:19:19
894	23	13	19	0	2025-03-18 20:19:20	2025-03-18 20:19:20
896	23	13	8	0	2025-03-18 20:19:20	2025-03-18 20:19:20
897	23	9	3	0	2025-03-18 20:19:33	2025-03-18 20:19:33
898	23	9	5	0	2025-03-18 20:19:33	2025-03-18 20:19:33
900	23	9	6	0	2025-03-18 20:19:33	2025-03-18 20:19:33
904	23	9	8	0	2025-03-18 20:19:34	2025-03-18 20:19:34
905	23	7	3	0	2025-03-18 20:19:40	2025-03-18 20:19:40
906	23	7	5	0	2025-03-18 20:19:40	2025-03-18 20:19:40
907	23	7	4	0	2025-03-18 20:19:40	2025-03-18 20:19:40
853	23	4	2	10	2025-03-18 20:18:46	2025-03-18 20:20:16
863	23	8	1	2	2025-03-18 20:18:52	2025-03-18 20:20:35
861	23	8	2	4	2025-03-18 20:18:52	2025-03-18 20:20:35
887	23	10	1	2	2025-03-18 20:19:12	2025-03-18 20:20:55
885	23	10	2	4	2025-03-18 20:19:12	2025-03-18 20:20:56
903	23	9	1	8	2025-03-18 20:19:34	2025-03-18 20:21:12
901	23	9	2	3	2025-03-18 20:19:34	2025-03-18 20:21:13
899	23	9	4	1	2025-03-18 20:19:33	2025-03-18 20:21:14
902	23	9	19	1	2025-03-18 20:19:34	2025-03-18 20:21:15
879	23	11	1	6	2025-03-18 20:19:08	2025-03-18 20:21:30
877	23	11	2	15	2025-03-18 20:19:07	2025-03-18 20:21:31
873	23	11	3	3	2025-03-18 20:19:06	2025-03-18 20:21:31
895	23	13	1	8	2025-03-18 20:19:20	2025-03-18 20:21:43
893	23	13	2	5	2025-03-18 20:19:19	2025-03-18 20:21:44
871	23	3	1	3	2025-03-18 20:19:02	2025-03-18 20:21:52
869	23	3	2	3	2025-03-18 20:19:01	2025-03-18 20:21:52
908	23	7	6	0	2025-03-18 20:19:41	2025-03-18 20:19:41
910	23	7	19	0	2025-03-18 20:19:42	2025-03-18 20:19:42
912	23	7	8	0	2025-03-18 20:19:43	2025-03-18 20:19:43
913	23	12	3	0	2025-03-18 20:20:05	2025-03-18 20:20:05
914	23	12	5	0	2025-03-18 20:20:06	2025-03-18 20:20:06
915	23	12	4	0	2025-03-18 20:20:06	2025-03-18 20:20:06
916	23	12	6	0	2025-03-18 20:20:06	2025-03-18 20:20:06
918	23	12	19	0	2025-03-18 20:20:06	2025-03-18 20:20:06
920	23	12	8	0	2025-03-18 20:20:07	2025-03-18 20:20:07
855	23	4	1	9	2025-03-18 20:18:46	2025-03-18 20:20:16
911	23	7	1	8	2025-03-18 20:19:42	2025-03-18 20:20:26
909	23	7	2	5	2025-03-18 20:19:41	2025-03-18 20:20:27
919	23	12	1	4	2025-03-18 20:20:07	2025-03-18 20:20:44
917	23	12	2	5	2025-03-18 20:20:06	2025-03-18 20:20:44
921	24	3	3	0	2025-03-18 20:24:21	2025-03-18 20:24:21
922	24	3	5	0	2025-03-18 20:24:21	2025-03-18 20:24:21
923	24	3	4	0	2025-03-18 20:24:22	2025-03-18 20:24:22
924	24	3	6	0	2025-03-18 20:24:22	2025-03-18 20:24:22
926	24	3	19	0	2025-03-18 20:24:22	2025-03-18 20:24:22
927	24	3	1	0	2025-03-18 20:24:23	2025-03-18 20:24:23
928	24	3	8	0	2025-03-18 20:24:23	2025-03-18 20:24:23
929	24	1	3	0	2025-03-18 20:24:28	2025-03-18 20:24:28
930	24	1	5	0	2025-03-18 20:24:28	2025-03-18 20:24:28
931	24	1	4	0	2025-03-18 20:24:28	2025-03-18 20:24:28
932	24	1	6	0	2025-03-18 20:24:28	2025-03-18 20:24:28
934	24	1	19	0	2025-03-18 20:24:28	2025-03-18 20:24:28
936	24	1	8	0	2025-03-18 20:24:28	2025-03-18 20:24:28
937	24	2	3	0	2025-03-18 20:24:32	2025-03-18 20:24:32
938	24	2	5	0	2025-03-18 20:24:32	2025-03-18 20:24:32
939	24	2	4	0	2025-03-18 20:24:32	2025-03-18 20:24:32
940	24	2	6	0	2025-03-18 20:24:32	2025-03-18 20:24:32
942	24	2	19	0	2025-03-18 20:24:33	2025-03-18 20:24:33
944	24	2	8	0	2025-03-18 20:24:33	2025-03-18 20:24:33
945	24	4	3	0	2025-03-18 20:24:37	2025-03-18 20:24:37
946	24	4	5	0	2025-03-18 20:24:37	2025-03-18 20:24:37
947	24	4	4	0	2025-03-18 20:24:37	2025-03-18 20:24:37
948	24	4	6	0	2025-03-18 20:24:37	2025-03-18 20:24:37
950	24	4	19	0	2025-03-18 20:24:38	2025-03-18 20:24:38
952	24	4	8	0	2025-03-18 20:24:38	2025-03-18 20:24:38
953	24	7	3	0	2025-03-18 20:24:44	2025-03-18 20:24:44
954	24	7	5	0	2025-03-18 20:24:44	2025-03-18 20:24:44
955	24	7	4	0	2025-03-18 20:24:44	2025-03-18 20:24:44
956	24	7	6	0	2025-03-18 20:24:44	2025-03-18 20:24:44
960	24	7	8	0	2025-03-18 20:24:44	2025-03-18 20:24:44
962	24	8	5	0	2025-03-18 20:24:49	2025-03-18 20:24:49
963	24	8	4	0	2025-03-18 20:24:50	2025-03-18 20:24:50
964	24	8	6	0	2025-03-18 20:24:50	2025-03-18 20:24:50
966	24	8	19	0	2025-03-18 20:24:50	2025-03-18 20:24:50
968	24	8	8	0	2025-03-18 20:24:51	2025-03-18 20:24:51
951	24	4	1	7	2025-03-18 20:24:38	2025-03-18 20:25:14
949	24	4	2	5	2025-03-18 20:24:38	2025-03-18 20:25:14
959	24	7	1	3	2025-03-18 20:24:44	2025-03-18 20:25:23
957	24	7	2	4	2025-03-18 20:24:44	2025-03-18 20:25:24
958	24	7	19	1	2025-03-18 20:24:44	2025-03-18 20:25:25
967	24	8	1	5	2025-03-18 20:24:50	2025-03-18 20:25:36
965	24	8	2	5	2025-03-18 20:24:50	2025-03-18 20:25:36
961	24	8	3	1	2025-03-18 20:24:49	2025-03-18 20:25:36
943	24	2	1	3	2025-03-18 20:24:33	2025-03-18 20:26:06
941	24	2	2	5	2025-03-18 20:24:33	2025-03-18 20:26:06
935	24	1	1	5	2025-03-18 20:24:28	2025-03-18 20:26:16
933	24	1	2	8	2025-03-18 20:24:28	2025-03-18 20:26:17
925	24	3	2	1	2025-03-18 20:24:22	2025-03-18 20:26:27
1231	31	4	4	0	2025-03-21 16:38:53	2025-03-21 16:38:53
1232	31	4	5	0	2025-03-21 16:38:53	2025-03-21 16:38:53
1233	31	4	6	0	2025-03-21 16:38:53	2025-03-21 16:38:53
1234	31	4	8	0	2025-03-21 16:38:53	2025-03-21 16:38:53
1235	31	4	19	0	2025-03-21 16:38:53	2025-03-21 16:38:53
1238	31	8	3	0	2025-03-21 16:38:59	2025-03-21 16:38:59
1239	31	8	4	0	2025-03-21 16:38:59	2025-03-21 16:38:59
1240	31	8	5	0	2025-03-21 16:39:00	2025-03-21 16:39:00
1241	31	8	6	0	2025-03-21 16:39:00	2025-03-21 16:39:00
1242	31	8	8	0	2025-03-21 16:39:00	2025-03-21 16:39:00
1243	31	8	19	0	2025-03-21 16:39:01	2025-03-21 16:39:01
1246	31	12	3	0	2025-03-21 16:39:05	2025-03-21 16:39:05
1247	31	12	4	0	2025-03-21 16:39:06	2025-03-21 16:39:06
1248	31	12	5	0	2025-03-21 16:39:06	2025-03-21 16:39:06
1249	31	12	6	0	2025-03-21 16:39:07	2025-03-21 16:39:07
1250	31	12	8	0	2025-03-21 16:39:07	2025-03-21 16:39:07
1251	31	12	19	0	2025-03-21 16:39:07	2025-03-21 16:39:07
1254	31	9	3	0	2025-03-21 16:39:11	2025-03-21 16:39:11
1255	31	9	4	0	2025-03-21 16:39:11	2025-03-21 16:39:11
1256	31	9	5	0	2025-03-21 16:39:11	2025-03-21 16:39:11
1257	31	9	6	0	2025-03-21 16:39:12	2025-03-21 16:39:12
1258	31	9	8	0	2025-03-21 16:39:12	2025-03-21 16:39:12
1259	31	9	19	0	2025-03-21 16:39:12	2025-03-21 16:39:12
1262	31	2	3	0	2025-03-21 16:39:16	2025-03-21 16:39:16
1263	31	2	4	0	2025-03-21 16:39:16	2025-03-21 16:39:16
1264	31	2	5	0	2025-03-21 16:39:17	2025-03-21 16:39:17
1265	31	2	6	0	2025-03-21 16:39:17	2025-03-21 16:39:17
1266	31	2	8	0	2025-03-21 16:39:17	2025-03-21 16:39:17
1267	31	2	19	0	2025-03-21 16:39:17	2025-03-21 16:39:17
1271	31	1	4	0	2025-03-21 16:39:21	2025-03-21 16:39:21
1272	31	1	5	0	2025-03-21 16:39:21	2025-03-21 16:39:21
1273	31	1	6	0	2025-03-21 16:39:21	2025-03-21 16:39:21
1274	31	1	8	0	2025-03-21 16:39:21	2025-03-21 16:39:21
1275	31	1	19	0	2025-03-21 16:39:21	2025-03-21 16:39:21
1278	31	11	3	0	2025-03-21 16:39:25	2025-03-21 16:39:25
1279	31	11	4	0	2025-03-21 16:39:25	2025-03-21 16:39:25
1280	31	11	5	0	2025-03-21 16:39:25	2025-03-21 16:39:25
1281	31	11	6	0	2025-03-21 16:39:26	2025-03-21 16:39:26
1282	31	11	8	0	2025-03-21 16:39:26	2025-03-21 16:39:26
1283	31	11	19	0	2025-03-21 16:39:26	2025-03-21 16:39:26
1286	31	13	3	0	2025-03-21 16:39:30	2025-03-21 16:39:30
1287	31	13	4	0	2025-03-21 16:39:30	2025-03-21 16:39:30
1288	31	13	5	0	2025-03-21 16:39:30	2025-03-21 16:39:30
1289	31	13	6	0	2025-03-21 16:39:31	2025-03-21 16:39:31
1290	31	13	8	0	2025-03-21 16:39:31	2025-03-21 16:39:31
1291	31	13	19	0	2025-03-21 16:39:31	2025-03-21 16:39:31
1292	31	3	1	0	2025-03-21 16:39:35	2025-03-21 16:39:35
1294	31	3	3	0	2025-03-21 16:39:36	2025-03-21 16:39:36
1295	31	3	4	0	2025-03-21 16:39:36	2025-03-21 16:39:36
1296	31	3	5	0	2025-03-21 16:39:36	2025-03-21 16:39:36
1297	31	3	6	0	2025-03-21 16:39:36	2025-03-21 16:39:36
1298	31	3	8	0	2025-03-21 16:39:36	2025-03-21 16:39:36
1299	31	3	19	0	2025-03-21 16:39:36	2025-03-21 16:39:36
1302	31	10	3	0	2025-03-21 16:39:47	2025-03-21 16:39:47
1303	31	10	4	0	2025-03-21 16:39:47	2025-03-21 16:39:47
1304	31	10	5	0	2025-03-21 16:39:48	2025-03-21 16:39:48
1305	31	10	6	0	2025-03-21 16:39:48	2025-03-21 16:39:48
1306	31	10	8	0	2025-03-21 16:39:48	2025-03-21 16:39:48
1307	31	10	19	0	2025-03-21 16:39:48	2025-03-21 16:39:48
1300	31	10	1	2	2025-03-21 16:39:47	2025-03-21 16:40:00
1301	31	10	2	3	2025-03-21 16:39:47	2025-03-21 16:40:01
1236	31	8	1	7	2025-03-21 16:38:59	2025-03-21 16:40:15
1237	31	8	2	11	2025-03-21 16:38:59	2025-03-21 16:40:16
1244	31	12	1	1	2025-03-21 16:39:05	2025-03-21 16:40:24
1245	31	12	2	4	2025-03-21 16:39:05	2025-03-21 16:40:24
1228	31	4	1	9	2025-03-21 16:38:53	2025-03-21 16:40:38
1229	31	4	2	7	2025-03-21 16:38:53	2025-03-21 16:40:40
1230	31	4	3	1	2025-03-21 16:38:53	2025-03-21 16:40:41
1252	31	9	1	7	2025-03-21 16:39:11	2025-03-21 16:40:52
1253	31	9	2	4	2025-03-21 16:39:11	2025-03-21 16:40:53
1284	31	13	1	6	2025-03-21 16:39:29	2025-03-21 16:41:07
1285	31	13	2	3	2025-03-21 16:39:30	2025-03-21 16:41:08
1276	31	11	1	6	2025-03-21 16:39:25	2025-03-21 16:41:29
1277	31	11	2	12	2025-03-21 16:39:25	2025-03-21 16:41:30
1260	31	2	1	6	2025-03-21 16:39:16	2025-03-21 16:41:55
1261	31	2	2	6	2025-03-21 16:39:16	2025-03-21 16:41:55
1268	31	1	1	5	2025-03-21 16:39:21	2025-03-21 16:42:04
1269	31	1	2	2	2025-03-21 16:39:21	2025-03-21 16:42:04
1270	31	1	3	1	2025-03-21 16:39:21	2025-03-21 16:42:05
1293	31	3	2	1	2025-03-21 16:39:36	2025-03-21 16:42:11
1498	35	33	31	0	2025-03-27 15:26:17	2025-03-27 15:26:17
1500	35	33	32	0	2025-03-27 15:26:17	2025-03-27 15:26:17
1501	35	33	33	0	2025-03-27 15:26:17	2025-03-27 15:26:17
1503	35	33	35	0	2025-03-27 15:26:17	2025-03-27 15:26:17
1494	35	33	29	7	2025-03-27 15:26:17	2025-03-27 15:39:03
1496	35	33	30	5	2025-03-27 15:26:17	2025-03-27 15:39:03
1502	35	33	34	1	2025-03-27 15:26:17	2025-03-27 15:39:03
1552	45	36	35	0	2025-04-25 12:06:14	2025-04-25 12:06:14
1553	45	32	29	10	2025-04-25 12:07:48	2025-04-25 12:07:48
1554	45	32	30	11	2025-04-25 12:07:49	2025-04-25 12:07:49
1555	45	32	31	1	2025-04-25 12:07:49	2025-04-25 12:07:49
1556	45	32	32	3	2025-04-25 12:07:49	2025-04-25 12:07:49
1557	45	32	33	0	2025-04-25 12:07:49	2025-04-25 12:07:49
1558	45	32	34	0	2025-04-25 12:07:50	2025-04-25 12:07:50
1448	34	34	29	2	2025-03-25 12:35:17	2025-03-25 12:36:29
1449	34	34	30	11	2025-03-25 12:35:18	2025-03-25 12:36:29
1541	36	33	31	0	2025-03-27 15:28:02	2025-03-27 15:28:02
1543	36	33	33	0	2025-03-27 15:28:03	2025-03-27 15:28:03
1544	36	33	34	0	2025-03-27 15:28:03	2025-03-27 15:28:03
1545	36	33	35	0	2025-03-27 15:28:03	2025-03-27 15:28:03
1420	34	36	29	7	2025-03-25 12:35:03	2025-03-27 15:34:42
1421	34	36	30	5	2025-03-25 12:35:03	2025-03-27 15:34:42
1423	34	36	32	2	2025-03-25 12:35:03	2025-03-27 15:34:43
1427	34	32	29	8	2025-03-25 12:35:09	2025-03-27 15:35:26
1428	34	32	30	8	2025-03-25 12:35:09	2025-03-27 15:35:26
1430	34	32	32	1	2025-03-25 12:35:10	2025-03-27 15:35:27
1434	34	33	29	2	2025-03-25 12:35:12	2025-03-27 15:35:39
1435	34	33	30	8	2025-03-25 12:35:12	2025-03-27 15:35:39
1441	34	35	29	8	2025-03-25 12:35:15	2025-03-27 15:36:03
1442	34	35	30	8	2025-03-25 12:35:15	2025-03-27 15:36:03
1455	34	37	29	4	2025-03-25 12:35:22	2025-03-27 15:36:34
1456	34	37	30	4	2025-03-25 12:35:23	2025-03-27 15:36:34
1458	34	37	32	1	2025-03-25 12:35:23	2025-03-27 15:36:35
1559	45	32	35	0	2025-04-25 12:07:50	2025-04-25 12:07:50
1560	45	33	29	10	2025-04-25 12:08:01	2025-04-25 12:08:01
1561	45	33	30	8	2025-04-25 12:08:01	2025-04-25 12:08:01
1562	45	33	31	0	2025-04-25 12:08:01	2025-04-25 12:08:01
1563	45	33	32	0	2025-04-25 12:08:01	2025-04-25 12:08:01
1564	45	33	33	0	2025-04-25 12:08:01	2025-04-25 12:08:01
1565	45	33	34	0	2025-04-25 12:08:01	2025-04-25 12:08:01
1566	45	33	35	0	2025-04-25 12:08:01	2025-04-25 12:08:01
1567	45	34	29	5	2025-04-25 12:08:13	2025-04-25 12:08:13
1568	45	34	30	5	2025-04-25 12:08:13	2025-04-25 12:08:13
1569	45	34	31	0	2025-04-25 12:08:13	2025-04-25 12:08:13
1570	45	34	32	0	2025-04-25 12:08:14	2025-04-25 12:08:14
1571	45	34	33	0	2025-04-25 12:08:14	2025-04-25 12:08:14
1572	45	34	34	1	2025-04-25 12:08:14	2025-04-25 12:08:14
1573	45	34	35	0	2025-04-25 12:08:14	2025-04-25 12:08:14
1574	45	37	29	2	2025-04-25 12:08:21	2025-04-25 12:08:21
1575	45	37	30	2	2025-04-25 12:08:22	2025-04-25 12:08:22
1576	45	37	31	0	2025-04-25 12:08:22	2025-04-25 12:08:22
1577	45	37	32	0	2025-04-25 12:08:22	2025-04-25 12:08:22
1578	45	37	33	0	2025-04-25 12:08:22	2025-04-25 12:08:22
1579	45	37	34	0	2025-04-25 12:08:22	2025-04-25 12:08:22
1580	45	37	35	0	2025-04-25 12:08:23	2025-04-25 12:08:23
1581	45	35	29	8	2025-04-25 12:08:36	2025-04-25 12:08:36
1582	45	35	30	7	2025-04-25 12:08:36	2025-04-25 12:08:36
1583	45	35	31	0	2025-04-25 12:08:36	2025-04-25 12:08:36
1584	45	35	32	3	2025-04-25 12:08:36	2025-04-25 12:08:36
1585	45	35	33	0	2025-04-25 12:08:36	2025-04-25 12:08:36
1586	45	35	34	0	2025-04-25 12:08:36	2025-04-25 12:08:36
1587	45	35	35	0	2025-04-25 12:08:36	2025-04-25 12:08:36
1588	45	31	29	8	2025-04-25 12:09:09	2025-04-25 12:09:09
1589	45	31	30	5	2025-04-25 12:09:09	2025-04-25 12:09:09
1590	45	31	31	0	2025-04-25 12:09:10	2025-04-25 12:09:10
1591	45	31	32	2	2025-04-25 12:09:10	2025-04-25 12:09:10
1592	45	31	33	0	2025-04-25 12:09:10	2025-04-25 12:09:10
1593	45	31	34	0	2025-04-25 12:09:11	2025-04-25 12:09:11
1594	45	31	35	0	2025-04-25 12:09:11	2025-04-25 12:09:11
1595	44	35	29	7	2025-04-25 12:09:37	2025-04-25 12:09:37
1596	44	35	30	11	2025-04-25 12:09:37	2025-04-25 12:09:37
1597	44	35	31	0	2025-04-25 12:09:37	2025-04-25 12:09:37
1598	44	35	32	0	2025-04-25 12:09:37	2025-04-25 12:09:37
1415	34	30	31	0	2025-03-25 12:34:59	2025-03-25 12:34:59
1416	34	30	32	0	2025-03-25 12:34:59	2025-03-25 12:34:59
1417	34	30	33	0	2025-03-25 12:34:59	2025-03-25 12:34:59
1418	34	30	34	0	2025-03-25 12:34:59	2025-03-25 12:34:59
1419	34	30	35	0	2025-03-25 12:34:59	2025-03-25 12:34:59
1422	34	36	31	0	2025-03-25 12:35:03	2025-03-25 12:35:03
1424	34	36	33	0	2025-03-25 12:35:03	2025-03-25 12:35:03
1425	34	36	34	0	2025-03-25 12:35:04	2025-03-25 12:35:04
1426	34	36	35	0	2025-03-25 12:35:04	2025-03-25 12:35:04
1429	34	32	31	0	2025-03-25 12:35:10	2025-03-25 12:35:10
1431	34	32	33	0	2025-03-25 12:35:10	2025-03-25 12:35:10
1432	34	32	34	0	2025-03-25 12:35:10	2025-03-25 12:35:10
1433	34	32	35	0	2025-03-25 12:35:10	2025-03-25 12:35:10
1436	34	33	31	0	2025-03-25 12:35:12	2025-03-25 12:35:12
1437	34	33	32	0	2025-03-25 12:35:12	2025-03-25 12:35:12
1438	34	33	33	0	2025-03-25 12:35:12	2025-03-25 12:35:12
1439	34	33	34	0	2025-03-25 12:35:12	2025-03-25 12:35:12
1440	34	33	35	0	2025-03-25 12:35:12	2025-03-25 12:35:12
1443	34	35	31	0	2025-03-25 12:35:15	2025-03-25 12:35:15
1444	34	35	32	0	2025-03-25 12:35:16	2025-03-25 12:35:16
1445	34	35	33	0	2025-03-25 12:35:16	2025-03-25 12:35:16
1446	34	35	34	0	2025-03-25 12:35:16	2025-03-25 12:35:16
1447	34	35	35	0	2025-03-25 12:35:16	2025-03-25 12:35:16
1450	34	34	31	0	2025-03-25 12:35:18	2025-03-25 12:35:18
1451	34	34	32	0	2025-03-25 12:35:18	2025-03-25 12:35:18
1452	34	34	33	0	2025-03-25 12:35:18	2025-03-25 12:35:18
1453	34	34	34	0	2025-03-25 12:35:18	2025-03-25 12:35:18
1454	34	34	35	0	2025-03-25 12:35:19	2025-03-25 12:35:19
1457	34	37	31	0	2025-03-25 12:35:23	2025-03-25 12:35:23
1459	34	37	33	0	2025-03-25 12:35:23	2025-03-25 12:35:23
1460	34	37	34	0	2025-03-25 12:35:23	2025-03-25 12:35:23
1461	34	37	35	0	2025-03-25 12:35:23	2025-03-25 12:35:23
1413	34	30	29	6	2025-03-25 12:34:59	2025-03-25 12:35:56
1414	34	30	30	8	2025-03-25 12:34:59	2025-03-25 12:35:56
1599	44	35	33	0	2025-04-25 12:09:37	2025-04-25 12:09:37
1600	44	35	34	1	2025-04-25 12:09:38	2025-04-25 12:09:38
1601	44	35	35	0	2025-04-25 12:09:38	2025-04-25 12:09:38
1602	44	31	29	6	2025-04-25 12:09:53	2025-04-25 12:09:53
1603	44	31	30	8	2025-04-25 12:09:53	2025-04-25 12:09:53
1604	44	31	31	0	2025-04-25 12:09:53	2025-04-25 12:09:53
1605	44	31	32	1	2025-04-25 12:09:53	2025-04-25 12:09:53
1606	44	31	33	0	2025-04-25 12:09:53	2025-04-25 12:09:53
1607	44	31	34	0	2025-04-25 12:09:53	2025-04-25 12:09:53
1608	44	31	35	0	2025-04-25 12:09:53	2025-04-25 12:09:53
1609	44	32	29	4	2025-04-25 12:10:06	2025-04-25 12:10:06
1610	44	32	30	5	2025-04-25 12:10:06	2025-04-25 12:10:06
1611	44	32	31	0	2025-04-25 12:10:06	2025-04-25 12:10:06
1612	44	32	32	1	2025-04-25 12:10:07	2025-04-25 12:10:07
1613	44	32	33	0	2025-04-25 12:10:07	2025-04-25 12:10:07
1614	44	32	34	0	2025-04-25 12:10:07	2025-04-25 12:10:07
1615	44	32	35	0	2025-04-25 12:10:07	2025-04-25 12:10:07
1616	44	36	29	5	2025-04-25 12:10:19	2025-04-25 12:10:19
1617	44	36	30	8	2025-04-25 12:10:19	2025-04-25 12:10:19
1618	44	36	31	0	2025-04-25 12:10:19	2025-04-25 12:10:19
1619	44	36	32	1	2025-04-25 12:10:20	2025-04-25 12:10:20
1620	44	36	33	0	2025-04-25 12:10:20	2025-04-25 12:10:20
1621	44	36	34	0	2025-04-25 12:10:20	2025-04-25 12:10:20
1622	44	36	35	0	2025-04-25 12:10:20	2025-04-25 12:10:20
1623	44	37	29	3	2025-04-25 12:10:28	2025-04-25 12:10:28
1624	44	37	30	3	2025-04-25 12:10:28	2025-04-25 12:10:28
1625	44	37	31	0	2025-04-25 12:10:28	2025-04-25 12:10:28
1626	44	37	32	0	2025-04-25 12:10:28	2025-04-25 12:10:28
1627	44	37	33	0	2025-04-25 12:10:29	2025-04-25 12:10:29
1628	44	37	34	0	2025-04-25 12:10:29	2025-04-25 12:10:29
1629	44	37	35	0	2025-04-25 12:10:29	2025-04-25 12:10:29
1542	36	33	32	0	2025-03-27 15:28:02	2025-04-25 12:15:16
1630	38	30	29	3	2025-04-25 12:17:01	2025-04-25 12:17:01
1631	38	30	30	5	2025-04-25 12:17:01	2025-04-25 12:17:01
1632	38	30	31	0	2025-04-25 12:17:01	2025-04-25 12:17:01
1633	38	30	32	0	2025-04-25 12:17:01	2025-04-25 12:17:01
1634	38	30	33	1	2025-04-25 12:17:02	2025-04-25 12:17:02
1635	38	30	34	0	2025-04-25 12:17:02	2025-04-25 12:17:02
1636	38	30	35	0	2025-04-25 12:17:02	2025-04-25 12:17:02
1637	38	31	29	7	2025-04-25 12:17:10	2025-04-25 12:17:10
1638	38	31	30	6	2025-04-25 12:17:10	2025-04-25 12:17:10
1639	38	31	31	0	2025-04-25 12:17:11	2025-04-25 12:17:11
1640	38	31	32	0	2025-04-25 12:17:11	2025-04-25 12:17:11
1641	38	31	33	0	2025-04-25 12:17:11	2025-04-25 12:17:11
1642	38	31	34	0	2025-04-25 12:17:11	2025-04-25 12:17:11
1643	38	31	35	0	2025-04-25 12:17:12	2025-04-25 12:17:12
1644	38	32	29	7	2025-04-25 12:17:28	2025-04-25 12:17:28
1645	38	32	30	10	2025-04-25 12:17:28	2025-04-25 12:17:28
1646	38	32	31	0	2025-04-25 12:17:28	2025-04-25 12:17:28
1647	38	32	32	2	2025-04-25 12:17:28	2025-04-25 12:17:28
1648	38	32	33	0	2025-04-25 12:17:28	2025-04-25 12:17:28
1649	38	32	34	0	2025-04-25 12:17:28	2025-04-25 12:17:28
1650	38	32	35	0	2025-04-25 12:17:28	2025-04-25 12:17:28
1651	38	33	29	4	2025-04-25 12:17:38	2025-04-25 12:17:38
1652	38	33	30	5	2025-04-25 12:17:38	2025-04-25 12:17:38
1653	38	33	31	0	2025-04-25 12:17:38	2025-04-25 12:17:38
1654	38	33	32	0	2025-04-25 12:17:38	2025-04-25 12:17:38
1655	38	33	33	0	2025-04-25 12:17:39	2025-04-25 12:17:39
1656	38	33	34	0	2025-04-25 12:17:39	2025-04-25 12:17:39
1657	38	33	35	0	2025-04-25 12:17:39	2025-04-25 12:17:39
1658	38	35	29	4	2025-04-25 12:17:49	2025-04-25 12:17:49
1659	38	35	30	9	2025-04-25 12:17:50	2025-04-25 12:17:50
1660	38	35	31	0	2025-04-25 12:17:50	2025-04-25 12:17:50
1661	38	35	32	0	2025-04-25 12:17:50	2025-04-25 12:17:50
1662	38	35	33	0	2025-04-25 12:17:50	2025-04-25 12:17:50
1663	38	35	34	0	2025-04-25 12:17:50	2025-04-25 12:17:50
1664	38	35	35	0	2025-04-25 12:17:51	2025-04-25 12:17:51
1665	38	34	29	2	2025-04-25 12:18:02	2025-04-25 12:18:02
1666	38	34	30	0	2025-04-25 12:18:02	2025-04-25 12:18:02
1667	38	34	31	0	2025-04-25 12:18:03	2025-04-25 12:18:03
1668	38	34	32	0	2025-04-25 12:18:03	2025-04-25 12:18:03
1669	38	34	33	0	2025-04-25 12:18:03	2025-04-25 12:18:03
1670	38	34	34	1	2025-04-25 12:18:03	2025-04-25 12:18:03
1671	38	34	35	0	2025-04-25 12:18:03	2025-04-25 12:18:03
1672	38	37	29	2	2025-04-25 12:18:21	2025-04-25 12:18:21
1673	38	37	30	2	2025-04-25 12:18:21	2025-04-25 12:18:21
1674	38	37	31	0	2025-04-25 12:18:21	2025-04-25 12:18:21
1675	38	37	32	0	2025-04-25 12:18:21	2025-04-25 12:18:21
1676	38	37	33	0	2025-04-25 12:18:21	2025-04-25 12:18:21
1677	38	37	34	0	2025-04-25 12:18:21	2025-04-25 12:18:21
1678	38	37	35	0	2025-04-25 12:18:21	2025-04-25 12:18:21
1679	47	30	29	9	2025-04-25 12:19:55	2025-04-25 12:19:55
1680	47	30	30	5	2025-04-25 12:19:55	2025-04-25 12:19:55
1681	47	30	31	0	2025-04-25 12:19:55	2025-04-25 12:19:55
1682	47	30	32	0	2025-04-25 12:19:56	2025-04-25 12:19:56
1683	47	30	33	1	2025-04-25 12:19:56	2025-04-25 12:19:56
1684	47	30	34	0	2025-04-25 12:19:56	2025-04-25 12:19:56
1685	47	30	35	0	2025-04-25 12:19:56	2025-04-25 12:19:56
1686	47	30	36	0	2025-04-25 12:19:56	2025-04-25 12:19:56
1687	47	31	29	5	2025-04-25 12:20:09	2025-04-25 12:20:09
1688	47	31	30	5	2025-04-25 12:20:10	2025-04-25 12:20:10
1689	47	31	31	0	2025-04-25 12:20:10	2025-04-25 12:20:10
1690	47	31	32	1	2025-04-25 12:20:10	2025-04-25 12:20:10
1691	47	31	33	0	2025-04-25 12:20:10	2025-04-25 12:20:10
1692	47	31	34	1	2025-04-25 12:20:11	2025-04-25 12:20:11
1693	47	31	35	0	2025-04-25 12:20:11	2025-04-25 12:20:11
1694	47	31	36	0	2025-04-25 12:20:11	2025-04-25 12:20:11
1695	47	32	29	5	2025-04-25 12:20:25	2025-04-25 12:20:25
1696	47	32	30	8	2025-04-25 12:20:25	2025-04-25 12:20:25
1697	47	32	31	1	2025-04-25 12:20:25	2025-04-25 12:20:25
1698	47	32	32	3	2025-04-25 12:20:25	2025-04-25 12:20:25
1699	47	32	33	0	2025-04-25 12:20:25	2025-04-25 12:20:25
1700	47	32	34	0	2025-04-25 12:20:25	2025-04-25 12:20:25
1701	47	32	35	0	2025-04-25 12:20:25	2025-04-25 12:20:25
1702	47	32	36	0	2025-04-25 12:20:25	2025-04-25 12:20:25
1703	47	33	29	3	2025-04-25 12:20:35	2025-04-25 12:20:35
1704	47	33	30	4	2025-04-25 12:20:35	2025-04-25 12:20:35
1705	47	33	31	1	2025-04-25 12:20:35	2025-04-25 12:20:35
1706	47	33	32	0	2025-04-25 12:20:36	2025-04-25 12:20:36
1707	47	33	33	0	2025-04-25 12:20:36	2025-04-25 12:20:36
1708	47	33	34	0	2025-04-25 12:20:37	2025-04-25 12:20:37
1709	47	33	35	0	2025-04-25 12:20:37	2025-04-25 12:20:37
1710	47	33	36	0	2025-04-25 12:20:37	2025-04-25 12:20:37
1711	47	34	29	4	2025-04-25 12:20:50	2025-04-25 12:20:50
1712	47	34	30	3	2025-04-25 12:20:50	2025-04-25 12:20:50
1713	47	34	31	1	2025-04-25 12:20:50	2025-04-25 12:20:50
1714	47	34	32	0	2025-04-25 12:20:50	2025-04-25 12:20:50
1715	47	34	33	0	2025-04-25 12:20:50	2025-04-25 12:20:50
1716	47	34	34	0	2025-04-25 12:20:50	2025-04-25 12:20:50
1717	47	34	35	0	2025-04-25 12:20:50	2025-04-25 12:20:50
1718	47	34	36	0	2025-04-25 12:20:50	2025-04-25 12:20:50
1719	47	35	29	7	2025-04-25 12:21:05	2025-04-25 12:21:05
1720	47	35	30	3	2025-04-25 12:21:06	2025-04-25 12:21:06
1721	47	35	31	0	2025-04-25 12:21:06	2025-04-25 12:21:06
1722	47	35	32	1	2025-04-25 12:21:06	2025-04-25 12:21:06
1723	47	35	33	0	2025-04-25 12:21:06	2025-04-25 12:21:06
1724	47	35	34	0	2025-04-25 12:21:07	2025-04-25 12:21:07
1725	47	35	35	0	2025-04-25 12:21:07	2025-04-25 12:21:07
1726	47	35	36	0	2025-04-25 12:21:07	2025-04-25 12:21:07
1727	47	36	29	4	2025-04-25 12:21:24	2025-04-25 12:21:24
1728	47	36	30	6	2025-04-25 12:21:24	2025-04-25 12:21:24
1729	47	36	31	0	2025-04-25 12:21:24	2025-04-25 12:21:24
1730	47	36	32	1	2025-04-25 12:21:25	2025-04-25 12:21:25
1731	47	36	33	0	2025-04-25 12:21:25	2025-04-25 12:21:25
1732	47	36	34	0	2025-04-25 12:21:25	2025-04-25 12:21:25
1733	47	36	35	0	2025-04-25 12:21:26	2025-04-25 12:21:26
1734	47	36	36	0	2025-04-25 12:21:26	2025-04-25 12:21:26
1735	40	30	29	3	2025-04-25 12:22:33	2025-04-25 12:22:33
1736	40	30	30	8	2025-04-25 12:22:33	2025-04-25 12:22:33
1737	40	30	31	0	2025-04-25 12:22:33	2025-04-25 12:22:33
1738	40	30	32	0	2025-04-25 12:22:33	2025-04-25 12:22:33
1739	40	30	33	0	2025-04-25 12:22:33	2025-04-25 12:22:33
1740	40	30	34	0	2025-04-25 12:22:34	2025-04-25 12:22:34
1741	40	30	35	0	2025-04-25 12:22:34	2025-04-25 12:22:34
1742	40	36	29	4	2025-04-25 12:22:43	2025-04-25 12:22:43
1743	40	36	30	5	2025-04-25 12:22:43	2025-04-25 12:22:43
1744	40	36	31	0	2025-04-25 12:22:43	2025-04-25 12:22:43
1745	40	36	32	1	2025-04-25 12:22:43	2025-04-25 12:22:43
1746	40	36	33	0	2025-04-25 12:22:43	2025-04-25 12:22:43
1747	40	36	34	0	2025-04-25 12:22:43	2025-04-25 12:22:43
1748	40	36	35	0	2025-04-25 12:22:43	2025-04-25 12:22:43
1749	40	33	29	4	2025-04-25 12:22:55	2025-04-25 12:22:55
1750	40	33	30	3	2025-04-25 12:22:55	2025-04-25 12:22:55
1751	40	33	31	0	2025-04-25 12:22:56	2025-04-25 12:22:56
1752	40	33	32	0	2025-04-25 12:22:56	2025-04-25 12:22:56
1753	40	33	33	0	2025-04-25 12:22:56	2025-04-25 12:22:56
1754	40	33	34	1	2025-04-25 12:22:56	2025-04-25 12:22:56
1755	40	33	35	0	2025-04-25 12:22:56	2025-04-25 12:22:56
1756	40	35	29	4	2025-04-25 12:23:07	2025-04-25 12:23:07
1757	40	35	30	9	2025-04-25 12:23:07	2025-04-25 12:23:07
1758	40	35	31	0	2025-04-25 12:23:07	2025-04-25 12:23:07
1759	40	35	32	1	2025-04-25 12:23:07	2025-04-25 12:23:07
1760	40	35	33	0	2025-04-25 12:23:07	2025-04-25 12:23:07
1761	40	35	34	0	2025-04-25 12:23:07	2025-04-25 12:23:07
1762	40	35	35	0	2025-04-25 12:23:07	2025-04-25 12:23:07
1763	40	37	29	4	2025-04-25 12:23:14	2025-04-25 12:23:14
1764	40	37	30	3	2025-04-25 12:23:15	2025-04-25 12:23:15
1765	40	37	31	0	2025-04-25 12:23:15	2025-04-25 12:23:15
1766	40	37	32	0	2025-04-25 12:23:15	2025-04-25 12:23:15
1767	40	37	33	0	2025-04-25 12:23:15	2025-04-25 12:23:15
1768	40	37	34	0	2025-04-25 12:23:16	2025-04-25 12:23:16
1769	40	37	35	0	2025-04-25 12:23:16	2025-04-25 12:23:16
1770	40	31	29	6	2025-04-25 12:23:25	2025-04-25 12:23:25
1771	40	31	30	9	2025-04-25 12:23:25	2025-04-25 12:23:25
1772	40	31	31	0	2025-04-25 12:23:25	2025-04-25 12:23:25
1773	40	31	32	3	2025-04-25 12:23:26	2025-04-25 12:23:26
1774	40	31	33	0	2025-04-25 12:23:26	2025-04-25 12:23:26
1775	40	31	34	0	2025-04-25 12:23:26	2025-04-25 12:23:26
1776	40	31	35	0	2025-04-25 12:23:26	2025-04-25 12:23:26
1777	48	30	29	17	2025-04-25 12:25:42	2025-04-25 12:25:42
1778	48	30	30	18	2025-04-25 12:25:42	2025-04-25 12:25:42
1779	48	30	31	0	2025-04-25 12:25:43	2025-04-25 12:25:43
1780	48	30	32	0	2025-04-25 12:25:43	2025-04-25 12:25:43
1781	48	30	33	0	2025-04-25 12:25:43	2025-04-25 12:25:43
1782	48	30	34	0	2025-04-25 12:25:43	2025-04-25 12:25:43
1783	48	30	35	0	2025-04-25 12:25:43	2025-04-25 12:25:43
1784	48	30	36	0	2025-04-25 12:25:44	2025-04-25 12:25:44
1785	48	31	29	13	2025-04-25 12:25:59	2025-04-25 12:25:59
1786	48	31	30	13	2025-04-25 12:25:59	2025-04-25 12:25:59
1787	48	31	31	1	2025-04-25 12:25:59	2025-04-25 12:25:59
1788	48	31	32	4	2025-04-25 12:26:00	2025-04-25 12:26:00
1789	48	31	33	0	2025-04-25 12:26:00	2025-04-25 12:26:00
1790	48	31	34	0	2025-04-25 12:26:00	2025-04-25 12:26:00
1791	48	31	35	0	2025-04-25 12:26:00	2025-04-25 12:26:00
1792	48	31	36	0	2025-04-25 12:26:01	2025-04-25 12:26:01
1793	48	32	29	21	2025-04-25 12:26:15	2025-04-25 12:26:15
1794	48	32	30	20	2025-04-25 12:26:15	2025-04-25 12:26:15
1795	48	32	31	0	2025-04-25 12:26:15	2025-04-25 12:26:15
1796	48	32	32	7	2025-04-25 12:26:15	2025-04-25 12:26:15
1797	48	32	33	0	2025-04-25 12:26:15	2025-04-25 12:26:15
1798	48	32	34	0	2025-04-25 12:26:16	2025-04-25 12:26:16
1799	48	32	35	0	2025-04-25 12:26:16	2025-04-25 12:26:16
1800	48	32	36	0	2025-04-25 12:26:16	2025-04-25 12:26:16
1801	48	33	29	11	2025-04-25 12:26:26	2025-04-25 12:26:26
1802	48	33	30	10	2025-04-25 12:26:26	2025-04-25 12:26:26
1803	48	33	31	0	2025-04-25 12:26:26	2025-04-25 12:26:26
1804	48	33	32	0	2025-04-25 12:26:26	2025-04-25 12:26:26
1805	48	33	33	0	2025-04-25 12:26:26	2025-04-25 12:26:26
1806	48	33	34	0	2025-04-25 12:26:26	2025-04-25 12:26:26
1807	48	33	35	0	2025-04-25 12:26:26	2025-04-25 12:26:26
1808	48	33	36	0	2025-04-25 12:26:27	2025-04-25 12:26:27
1809	48	34	29	9	2025-04-25 12:26:42	2025-04-25 12:26:42
1810	48	34	30	9	2025-04-25 12:26:42	2025-04-25 12:26:42
1811	48	34	31	1	2025-04-25 12:26:43	2025-04-25 12:26:43
1812	48	34	32	0	2025-04-25 12:26:43	2025-04-25 12:26:43
1813	48	34	33	0	2025-04-25 12:26:43	2025-04-25 12:26:43
1814	48	34	34	0	2025-04-25 12:26:43	2025-04-25 12:26:43
1815	48	34	35	0	2025-04-25 12:26:44	2025-04-25 12:26:44
1816	48	34	36	0	2025-04-25 12:26:44	2025-04-25 12:26:44
1817	42	34	29	8	2025-04-25 12:28:06	2025-04-25 12:28:06
1818	42	34	30	5	2025-04-25 12:28:06	2025-04-25 12:28:06
1819	42	34	31	0	2025-04-25 12:28:06	2025-04-25 12:28:06
1820	42	34	32	0	2025-04-25 12:28:06	2025-04-25 12:28:06
1821	42	34	33	0	2025-04-25 12:28:07	2025-04-25 12:28:07
1822	42	34	34	0	2025-04-25 12:28:07	2025-04-25 12:28:07
1823	42	34	35	0	2025-04-25 12:28:07	2025-04-25 12:28:07
1824	42	32	29	8	2025-04-25 12:28:15	2025-04-25 12:28:15
1825	42	32	30	15	2025-04-25 12:28:15	2025-04-25 12:28:15
1826	42	32	31	0	2025-04-25 12:28:15	2025-04-25 12:28:15
1828	42	32	33	0	2025-04-25 12:28:16	2025-04-25 12:28:16
1829	42	32	34	0	2025-04-25 12:28:16	2025-04-25 12:28:16
1830	42	32	35	0	2025-04-25 12:28:16	2025-04-25 12:28:16
1827	42	32	32	5	2025-04-25 12:28:16	2025-04-25 12:28:21
1831	42	37	29	4	2025-04-25 12:28:27	2025-04-25 12:28:27
1832	42	37	30	1	2025-04-25 12:28:27	2025-04-25 12:28:27
1833	42	37	31	0	2025-04-25 12:28:27	2025-04-25 12:28:27
1834	42	37	32	0	2025-04-25 12:28:27	2025-04-25 12:28:27
1835	42	37	33	0	2025-04-25 12:28:27	2025-04-25 12:28:27
1836	42	37	34	0	2025-04-25 12:28:27	2025-04-25 12:28:27
1837	42	37	35	0	2025-04-25 12:28:27	2025-04-25 12:28:27
1838	42	35	29	15	2025-04-25 12:28:38	2025-04-25 12:28:38
1839	42	35	30	12	2025-04-25 12:28:38	2025-04-25 12:28:38
1840	42	35	31	0	2025-04-25 12:28:39	2025-04-25 12:28:39
1841	42	35	32	0	2025-04-25 12:28:39	2025-04-25 12:28:39
1842	42	35	33	0	2025-04-25 12:28:39	2025-04-25 12:28:39
1843	42	35	34	0	2025-04-25 12:28:39	2025-04-25 12:28:39
1844	42	35	35	0	2025-04-25 12:28:39	2025-04-25 12:28:39
1845	43	30	29	8	2025-04-25 12:29:48	2025-04-25 12:29:48
1846	43	30	30	7	2025-04-25 12:29:48	2025-04-25 12:29:48
1847	43	30	31	0	2025-04-25 12:29:49	2025-04-25 12:29:49
1848	43	30	32	0	2025-04-25 12:29:49	2025-04-25 12:29:49
1849	43	30	33	0	2025-04-25 12:29:49	2025-04-25 12:29:49
1850	43	30	34	1	2025-04-25 12:29:49	2025-04-25 12:29:49
1851	43	30	35	0	2025-04-25 12:29:49	2025-04-25 12:29:49
1852	43	31	29	12	2025-04-25 12:30:05	2025-04-25 12:30:05
1853	43	31	30	5	2025-04-25 12:30:05	2025-04-25 12:30:05
1854	43	31	31	0	2025-04-25 12:30:05	2025-04-25 12:30:05
1855	43	31	32	2	2025-04-25 12:30:05	2025-04-25 12:30:05
1856	43	31	33	0	2025-04-25 12:30:05	2025-04-25 12:30:05
1857	43	31	34	0	2025-04-25 12:30:06	2025-04-25 12:30:06
1858	43	31	35	0	2025-04-25 12:30:06	2025-04-25 12:30:06
1859	43	32	29	7	2025-04-25 12:30:14	2025-04-25 12:30:14
1860	43	32	30	7	2025-04-25 12:30:14	2025-04-25 12:30:14
1861	43	32	31	0	2025-04-25 12:30:14	2025-04-25 12:30:14
1862	43	32	32	1	2025-04-25 12:30:14	2025-04-25 12:30:14
1863	43	32	33	0	2025-04-25 12:30:14	2025-04-25 12:30:14
1864	43	32	34	0	2025-04-25 12:30:15	2025-04-25 12:30:15
1865	43	32	35	0	2025-04-25 12:30:15	2025-04-25 12:30:15
1866	43	33	29	2	2025-04-25 12:30:21	2025-04-25 12:30:21
1867	43	33	30	2	2025-04-25 12:30:21	2025-04-25 12:30:21
1868	43	33	31	0	2025-04-25 12:30:22	2025-04-25 12:30:22
1869	43	33	32	0	2025-04-25 12:30:22	2025-04-25 12:30:22
1870	43	33	33	0	2025-04-25 12:30:22	2025-04-25 12:30:22
1871	43	33	34	0	2025-04-25 12:30:22	2025-04-25 12:30:22
1872	43	33	35	0	2025-04-25 12:30:22	2025-04-25 12:30:22
1873	43	36	29	0	2025-04-25 12:30:35	2025-04-25 12:30:35
1874	43	36	30	5	2025-04-25 12:30:35	2025-04-25 12:30:35
1875	43	36	31	0	2025-04-25 12:30:36	2025-04-25 12:30:36
1876	43	36	32	0	2025-04-25 12:30:36	2025-04-25 12:30:36
1877	43	36	33	0	2025-04-25 12:30:36	2025-04-25 12:30:36
1878	43	36	34	1	2025-04-25 12:30:36	2025-04-25 12:30:36
1879	43	36	35	1	2025-04-25 12:30:37	2025-04-25 12:30:37
1880	43	34	29	3	2025-04-25 12:31:45	2025-04-25 12:31:45
1881	43	34	30	5	2025-04-25 12:31:45	2025-04-25 12:31:45
1882	43	34	31	0	2025-04-25 12:31:45	2025-04-25 12:31:45
1883	43	34	32	1	2025-04-25 12:31:45	2025-04-25 12:31:45
1884	43	34	33	0	2025-04-25 12:31:45	2025-04-25 12:31:45
1885	43	34	34	0	2025-04-25 12:31:45	2025-04-25 12:31:45
1886	43	34	35	0	2025-04-25 12:31:46	2025-04-25 12:31:46
1887	43	35	29	7	2025-04-25 12:31:58	2025-04-25 12:31:58
1888	43	35	30	6	2025-04-25 12:31:58	2025-04-25 12:31:58
1889	43	35	31	0	2025-04-25 12:31:59	2025-04-25 12:31:59
1890	43	35	32	0	2025-04-25 12:31:59	2025-04-25 12:31:59
1891	43	35	33	0	2025-04-25 12:31:59	2025-04-25 12:31:59
1892	43	35	34	0	2025-04-25 12:32:00	2025-04-25 12:32:00
1893	43	35	35	0	2025-04-25 12:32:00	2025-04-25 12:32:00
1894	43	37	29	1	2025-04-25 12:32:09	2025-04-25 12:32:09
1895	43	37	30	2	2025-04-25 12:32:09	2025-04-25 12:32:09
1896	43	37	31	0	2025-04-25 12:32:09	2025-04-25 12:32:09
1897	43	37	32	0	2025-04-25 12:32:09	2025-04-25 12:32:09
1898	43	37	33	0	2025-04-25 12:32:09	2025-04-25 12:32:09
1899	43	37	34	0	2025-04-25 12:32:09	2025-04-25 12:32:09
1900	43	37	35	0	2025-04-25 12:32:09	2025-04-25 12:32:09
1941	50	35	45	0	2025-04-25 12:55:25	2025-04-25 12:55:25
1942	50	35	46	0	2025-04-25 12:55:25	2025-04-25 12:55:25
1943	50	35	47	0	2025-04-25 12:55:25	2025-04-25 12:55:25
1944	50	35	48	0	2025-04-25 12:55:26	2025-04-25 12:55:26
1945	50	35	49	7	2025-04-25 12:55:26	2025-04-25 12:55:26
1946	50	35	50	11	2025-04-25 12:55:26	2025-04-25 12:55:26
1947	50	35	51	0	2025-04-25 12:55:27	2025-04-25 12:55:27
1948	50	35	52	1	2025-04-25 12:55:27	2025-04-25 12:55:27
1949	50	31	45	0	2025-04-25 12:55:36	2025-04-25 12:55:36
1950	50	31	46	0	2025-04-25 12:55:37	2025-04-25 12:55:37
1951	50	31	47	0	2025-04-25 12:55:37	2025-04-25 12:55:37
1952	50	31	48	1	2025-04-25 12:55:37	2025-04-25 12:55:37
1953	50	31	49	6	2025-04-25 12:55:37	2025-04-25 12:55:37
1954	50	31	50	8	2025-04-25 12:55:37	2025-04-25 12:55:37
1955	50	31	51	0	2025-04-25 12:55:38	2025-04-25 12:55:38
1956	50	31	52	0	2025-04-25 12:55:38	2025-04-25 12:55:38
1957	50	32	45	0	2025-04-25 12:55:59	2025-04-25 12:55:59
1958	50	32	46	0	2025-04-25 12:56:00	2025-04-25 12:56:00
1959	50	32	47	0	2025-04-25 12:56:00	2025-04-25 12:56:00
1963	50	32	51	0	2025-04-25 12:56:01	2025-04-25 12:56:01
1964	50	32	52	0	2025-04-25 12:56:01	2025-04-25 12:56:01
1960	50	32	48	1	2025-04-25 12:56:00	2025-04-25 12:56:14
1961	50	32	49	4	2025-04-25 12:56:00	2025-04-25 12:56:14
1962	50	32	50	5	2025-04-25 12:56:01	2025-04-25 12:56:14
1965	50	36	45	0	2025-04-25 12:56:21	2025-04-25 12:56:21
1966	50	36	46	0	2025-04-25 12:56:21	2025-04-25 12:56:21
1967	50	36	47	0	2025-04-25 12:56:21	2025-04-25 12:56:21
1968	50	36	48	1	2025-04-25 12:56:21	2025-04-25 12:56:21
1969	50	36	49	5	2025-04-25 12:56:21	2025-04-25 12:56:21
1970	50	36	50	8	2025-04-25 12:56:22	2025-04-25 12:56:22
1971	50	36	51	0	2025-04-25 12:56:22	2025-04-25 12:56:22
1972	50	36	52	0	2025-04-25 12:56:22	2025-04-25 12:56:22
1973	50	37	45	0	2025-04-25 12:56:28	2025-04-25 12:56:28
1974	50	37	46	0	2025-04-25 12:56:29	2025-04-25 12:56:29
1975	50	37	47	0	2025-04-25 12:56:29	2025-04-25 12:56:29
1976	50	37	48	0	2025-04-25 12:56:29	2025-04-25 12:56:29
1977	50	37	49	3	2025-04-25 12:56:29	2025-04-25 12:56:29
1978	50	37	50	3	2025-04-25 12:56:30	2025-04-25 12:56:30
1979	50	37	51	0	2025-04-25 12:56:30	2025-04-25 12:56:30
1980	50	37	52	0	2025-04-25 12:56:30	2025-04-25 12:56:30
2045	53	10	1	3	2025-06-13 15:36:42	2025-06-13 15:36:42
2046	53	10	2	2	2025-06-13 15:36:42	2025-06-13 15:36:42
2047	53	10	3	0	2025-06-13 15:36:42	2025-06-13 15:36:42
2048	53	10	4	0	2025-06-13 15:36:42	2025-06-13 15:36:42
2049	53	10	5	0	2025-06-13 15:36:42	2025-06-13 15:36:42
2050	53	10	6	0	2025-06-13 15:36:42	2025-06-13 15:36:42
2051	53	10	8	0	2025-06-13 15:36:42	2025-06-13 15:36:42
2052	53	10	19	0	2025-06-13 15:36:42	2025-06-13 15:36:42
2053	53	8	1	4	2025-06-13 15:36:50	2025-06-13 15:36:50
2054	53	8	2	2	2025-06-13 15:36:51	2025-06-13 15:36:51
2055	53	8	3	2	2025-06-13 15:36:51	2025-06-13 15:36:51
2056	53	8	4	0	2025-06-13 15:36:51	2025-06-13 15:36:51
2057	53	8	5	0	2025-06-13 15:36:51	2025-06-13 15:36:51
2058	53	8	6	0	2025-06-13 15:36:51	2025-06-13 15:36:51
2059	53	8	8	0	2025-06-13 15:36:51	2025-06-13 15:36:51
2060	53	8	19	1	2025-06-13 15:36:51	2025-06-13 15:36:51
2061	53	4	1	7	2025-06-13 15:37:45	2025-06-13 15:37:45
2062	53	4	2	3	2025-06-13 15:37:46	2025-06-13 15:37:46
2063	53	4	3	2	2025-06-13 15:37:48	2025-06-13 15:37:48
2064	53	4	4	0	2025-06-13 15:37:49	2025-06-13 15:37:49
2065	53	4	5	0	2025-06-13 15:37:49	2025-06-13 15:37:49
2066	53	4	6	0	2025-06-13 15:37:50	2025-06-13 15:37:50
2067	53	4	8	0	2025-06-13 15:37:50	2025-06-13 15:37:50
2068	53	4	19	0	2025-06-13 15:37:51	2025-06-13 15:37:51
2069	53	9	1	5	2025-06-13 15:38:00	2025-06-13 15:38:00
2070	53	9	2	2	2025-06-13 15:38:01	2025-06-13 15:38:01
2071	53	9	3	0	2025-06-13 15:38:03	2025-06-13 15:38:03
2072	53	9	4	0	2025-06-13 15:38:04	2025-06-13 15:38:04
2073	53	9	5	0	2025-06-13 15:38:04	2025-06-13 15:38:04
2074	53	9	6	0	2025-06-13 15:38:05	2025-06-13 15:38:05
2075	53	9	8	0	2025-06-13 15:38:05	2025-06-13 15:38:05
2076	53	9	19	0	2025-06-13 15:38:05	2025-06-13 15:38:05
2077	53	2	1	5	2025-06-13 15:38:10	2025-06-13 15:38:10
2078	53	2	2	3	2025-06-13 15:38:11	2025-06-13 15:38:11
2079	53	2	3	0	2025-06-13 15:38:11	2025-06-13 15:38:11
2080	53	2	4	0	2025-06-13 15:38:11	2025-06-13 15:38:11
2081	53	2	5	0	2025-06-13 15:38:11	2025-06-13 15:38:11
2082	53	2	6	0	2025-06-13 15:38:11	2025-06-13 15:38:11
2083	53	2	8	0	2025-06-13 15:38:12	2025-06-13 15:38:12
2084	53	2	19	0	2025-06-13 15:38:12	2025-06-13 15:38:12
2085	53	7	1	2	2025-06-13 15:38:18	2025-06-13 15:38:18
2086	53	7	2	4	2025-06-13 15:38:19	2025-06-13 15:38:19
2087	53	7	3	0	2025-06-13 15:38:19	2025-06-13 15:38:19
2088	53	7	4	0	2025-06-13 15:38:19	2025-06-13 15:38:19
2089	53	7	5	0	2025-06-13 15:38:19	2025-06-13 15:38:19
2090	53	7	6	0	2025-06-13 15:38:19	2025-06-13 15:38:19
2091	53	7	8	0	2025-06-13 15:38:20	2025-06-13 15:38:20
2092	53	7	19	0	2025-06-13 15:38:21	2025-06-13 15:38:21
2093	53	1	1	4	2025-06-13 15:38:26	2025-06-13 15:38:26
2094	53	1	2	7	2025-06-13 15:38:27	2025-06-13 15:38:27
2095	53	1	3	0	2025-06-13 15:38:27	2025-06-13 15:38:27
2096	53	1	4	0	2025-06-13 15:38:27	2025-06-13 15:38:27
2097	53	1	5	0	2025-06-13 15:38:27	2025-06-13 15:38:27
2098	53	1	6	0	2025-06-13 15:38:27	2025-06-13 15:38:27
2099	53	1	8	0	2025-06-13 15:38:27	2025-06-13 15:38:27
2100	53	1	19	0	2025-06-13 15:38:27	2025-06-13 15:38:27
2101	53	3	1	2	2025-06-13 15:38:32	2025-06-13 15:38:32
2102	53	3	2	2	2025-06-13 15:38:33	2025-06-13 15:38:33
2103	53	3	3	0	2025-06-13 15:38:34	2025-06-13 15:38:34
2104	53	3	4	0	2025-06-13 15:38:35	2025-06-13 15:38:35
2105	53	3	5	0	2025-06-13 15:38:35	2025-06-13 15:38:35
2106	53	3	6	0	2025-06-13 15:38:36	2025-06-13 15:38:36
2107	53	3	8	0	2025-06-13 15:38:36	2025-06-13 15:38:36
2108	53	3	19	2	2025-06-13 15:38:36	2025-06-13 15:38:36
2109	54	10	1	4	2025-06-13 15:40:41	2025-06-13 15:40:41
2110	54	10	2	1	2025-06-13 15:40:43	2025-06-13 15:40:43
2111	54	10	3	0	2025-06-13 15:40:43	2025-06-13 15:40:43
2112	54	10	4	0	2025-06-13 15:40:44	2025-06-13 15:40:44
2113	54	10	5	0	2025-06-13 15:40:45	2025-06-13 15:40:45
2114	54	10	6	0	2025-06-13 15:40:46	2025-06-13 15:40:46
2115	54	10	8	0	2025-06-13 15:40:46	2025-06-13 15:40:46
2116	54	10	19	0	2025-06-13 15:40:47	2025-06-13 15:40:47
2117	54	8	1	3	2025-06-13 15:40:55	2025-06-13 15:40:55
2118	54	8	2	2	2025-06-13 15:40:55	2025-06-13 15:40:55
2119	54	8	3	2	2025-06-13 15:40:56	2025-06-13 15:40:56
2120	54	8	4	0	2025-06-13 15:40:58	2025-06-13 15:40:58
2121	54	8	5	0	2025-06-13 15:40:58	2025-06-13 15:40:58
2122	54	8	6	0	2025-06-13 15:40:58	2025-06-13 15:40:58
2123	54	8	8	0	2025-06-13 15:40:58	2025-06-13 15:40:58
2124	54	8	19	0	2025-06-13 15:40:58	2025-06-13 15:40:58
2125	54	4	1	8	2025-06-13 15:41:06	2025-06-13 15:41:06
2126	54	4	2	2	2025-06-13 15:41:06	2025-06-13 15:41:06
2127	54	4	3	0	2025-06-13 15:41:06	2025-06-13 15:41:06
2128	54	4	4	0	2025-06-13 15:41:06	2025-06-13 15:41:06
2129	54	4	5	0	2025-06-13 15:41:06	2025-06-13 15:41:06
2130	54	4	6	0	2025-06-13 15:41:06	2025-06-13 15:41:06
2131	54	4	8	0	2025-06-13 15:41:06	2025-06-13 15:41:06
2132	54	4	19	0	2025-06-13 15:41:06	2025-06-13 15:41:06
2133	54	9	1	7	2025-06-13 15:41:12	2025-06-13 15:41:12
2134	54	9	2	2	2025-06-13 15:41:12	2025-06-13 15:41:12
2135	54	9	3	0	2025-06-13 15:41:13	2025-06-13 15:41:13
2136	54	9	4	0	2025-06-13 15:41:14	2025-06-13 15:41:14
2137	54	9	5	0	2025-06-13 15:41:14	2025-06-13 15:41:14
2138	54	9	6	0	2025-06-13 15:41:14	2025-06-13 15:41:14
2139	54	9	8	0	2025-06-13 15:41:15	2025-06-13 15:41:15
2140	54	9	19	0	2025-06-13 15:41:15	2025-06-13 15:41:15
2141	54	11	1	7	2025-06-13 15:41:24	2025-06-13 15:41:24
2142	54	11	2	5	2025-06-13 15:41:24	2025-06-13 15:41:24
2143	54	11	3	0	2025-06-13 15:41:25	2025-06-13 15:41:25
2144	54	11	4	0	2025-06-13 15:41:25	2025-06-13 15:41:25
2145	54	11	5	1	2025-06-13 15:41:26	2025-06-13 15:41:26
2146	54	11	6	0	2025-06-13 15:41:27	2025-06-13 15:41:27
2147	54	11	8	0	2025-06-13 15:41:28	2025-06-13 15:41:28
2148	54	11	19	0	2025-06-13 15:41:28	2025-06-13 15:41:28
2149	54	2	1	6	2025-06-13 15:41:37	2025-06-13 15:41:37
2150	54	2	2	3	2025-06-13 15:41:39	2025-06-13 15:41:39
2151	54	2	3	0	2025-06-13 15:41:39	2025-06-13 15:41:39
2152	54	2	4	0	2025-06-13 15:41:39	2025-06-13 15:41:39
2153	54	2	5	0	2025-06-13 15:41:39	2025-06-13 15:41:39
2154	54	2	6	0	2025-06-13 15:41:39	2025-06-13 15:41:39
2155	54	2	8	0	2025-06-13 15:41:39	2025-06-13 15:41:39
2156	54	2	19	0	2025-06-13 15:41:39	2025-06-13 15:41:39
2157	54	7	1	2	2025-06-13 15:41:44	2025-06-13 15:41:44
2158	54	7	2	5	2025-06-13 15:41:46	2025-06-13 15:41:46
2159	54	7	3	0	2025-06-13 15:41:47	2025-06-13 15:41:47
2160	54	7	4	0	2025-06-13 15:41:47	2025-06-13 15:41:47
2161	54	7	5	0	2025-06-13 15:41:47	2025-06-13 15:41:47
2162	54	7	6	0	2025-06-13 15:41:47	2025-06-13 15:41:47
2163	54	7	8	0	2025-06-13 15:41:47	2025-06-13 15:41:47
2164	54	7	19	0	2025-06-13 15:41:47	2025-06-13 15:41:47
2165	54	1	1	3	2025-06-13 15:41:52	2025-06-13 15:41:52
2166	54	1	2	4	2025-06-13 15:41:52	2025-06-13 15:41:52
2167	54	1	3	0	2025-06-13 15:41:54	2025-06-13 15:41:54
2168	54	1	4	0	2025-06-13 15:41:55	2025-06-13 15:41:55
2169	54	1	5	0	2025-06-13 15:41:55	2025-06-13 15:41:55
2170	54	1	6	0	2025-06-13 15:41:55	2025-06-13 15:41:55
2171	54	1	8	0	2025-06-13 15:41:55	2025-06-13 15:41:55
2172	54	1	19	0	2025-06-13 15:41:55	2025-06-13 15:41:55
2173	54	3	1	0	2025-06-13 15:41:59	2025-06-13 15:41:59
2174	54	3	2	1	2025-06-13 15:41:59	2025-06-13 15:41:59
2175	54	3	3	0	2025-06-13 15:42:00	2025-06-13 15:42:00
2176	54	3	4	0	2025-06-13 15:42:01	2025-06-13 15:42:01
2177	54	3	5	0	2025-06-13 15:42:01	2025-06-13 15:42:01
2178	54	3	6	0	2025-06-13 15:42:01	2025-06-13 15:42:01
2179	54	3	8	0	2025-06-13 15:42:01	2025-06-13 15:42:01
2180	54	3	19	0	2025-06-13 15:42:01	2025-06-13 15:42:01
2181	55	10	1	2	2025-06-16 12:12:32	2025-06-16 12:12:32
2182	55	10	2	2	2025-06-16 12:12:32	2025-06-16 12:12:32
2183	55	10	3	0	2025-06-16 12:12:32	2025-06-16 12:12:32
2184	55	10	4	0	2025-06-16 12:12:32	2025-06-16 12:12:32
2185	55	10	5	0	2025-06-16 12:12:32	2025-06-16 12:12:32
2186	55	10	6	0	2025-06-16 12:12:32	2025-06-16 12:12:32
2187	55	10	8	0	2025-06-16 12:12:32	2025-06-16 12:12:32
2188	55	10	19	0	2025-06-16 12:12:32	2025-06-16 12:12:32
2189	55	8	1	4	2025-06-16 12:13:05	2025-06-16 12:13:05
2190	55	8	2	1	2025-06-16 12:13:06	2025-06-16 12:13:06
2191	55	8	3	0	2025-06-16 12:13:07	2025-06-16 12:13:07
2192	55	8	4	0	2025-06-16 12:13:08	2025-06-16 12:13:08
2193	55	8	5	0	2025-06-16 12:13:09	2025-06-16 12:13:09
2194	55	8	6	0	2025-06-16 12:13:09	2025-06-16 12:13:09
2195	55	8	8	0	2025-06-16 12:13:09	2025-06-16 12:13:09
2196	55	8	19	0	2025-06-16 12:13:09	2025-06-16 12:13:09
2197	55	4	1	8	2025-06-16 12:13:23	2025-06-16 12:13:23
2198	55	4	2	3	2025-06-16 12:13:24	2025-06-16 12:13:24
2199	55	4	3	0	2025-06-16 12:13:24	2025-06-16 12:13:24
2200	55	4	4	0	2025-06-16 12:13:24	2025-06-16 12:13:24
2201	55	4	5	0	2025-06-16 12:13:24	2025-06-16 12:13:24
2202	55	4	6	0	2025-06-16 12:13:24	2025-06-16 12:13:24
2203	55	4	8	0	2025-06-16 12:13:24	2025-06-16 12:13:24
2204	55	4	19	0	2025-06-16 12:13:24	2025-06-16 12:13:24
2205	55	9	1	5	2025-06-16 12:13:54	2025-06-16 12:13:54
2206	55	9	2	5	2025-06-16 12:13:55	2025-06-16 12:13:55
2207	55	9	3	0	2025-06-16 12:13:56	2025-06-16 12:13:56
2208	55	9	4	0	2025-06-16 12:13:57	2025-06-16 12:13:57
2209	55	9	5	1	2025-06-16 12:13:57	2025-06-16 12:13:57
2210	55	9	6	0	2025-06-16 12:13:57	2025-06-16 12:13:57
2211	55	9	8	0	2025-06-16 12:13:57	2025-06-16 12:13:57
2212	55	9	19	0	2025-06-16 12:13:57	2025-06-16 12:13:57
2213	55	11	1	4	2025-06-16 12:14:31	2025-06-16 12:14:31
2214	55	11	2	6	2025-06-16 12:14:32	2025-06-16 12:14:32
2215	55	11	3	0	2025-06-16 12:14:33	2025-06-16 12:14:33
2216	55	11	4	0	2025-06-16 12:14:34	2025-06-16 12:14:34
2217	55	11	5	0	2025-06-16 12:14:34	2025-06-16 12:14:34
2218	55	11	6	0	2025-06-16 12:14:34	2025-06-16 12:14:34
2219	55	11	8	0	2025-06-16 12:14:34	2025-06-16 12:14:34
2220	55	11	19	0	2025-06-16 12:14:34	2025-06-16 12:14:34
2221	55	2	1	2	2025-06-16 12:14:47	2025-06-16 12:14:47
2222	55	2	2	1	2025-06-16 12:14:48	2025-06-16 12:14:48
2223	55	2	3	0	2025-06-16 12:14:50	2025-06-16 12:14:50
2224	55	2	4	0	2025-06-16 12:14:52	2025-06-16 12:14:52
2225	55	2	5	0	2025-06-16 12:14:52	2025-06-16 12:14:52
2226	55	2	6	0	2025-06-16 12:14:52	2025-06-16 12:14:52
2227	55	2	8	0	2025-06-16 12:14:52	2025-06-16 12:14:52
2228	55	2	19	0	2025-06-16 12:14:52	2025-06-16 12:14:52
2229	55	1	1	5	2025-06-16 12:15:05	2025-06-16 12:15:05
2230	55	1	2	4	2025-06-16 12:15:07	2025-06-16 12:15:07
2231	55	1	3	0	2025-06-16 12:15:07	2025-06-16 12:15:07
2232	55	1	4	0	2025-06-16 12:15:07	2025-06-16 12:15:07
2233	55	1	5	0	2025-06-16 12:15:07	2025-06-16 12:15:07
2234	55	1	6	0	2025-06-16 12:15:07	2025-06-16 12:15:07
2235	55	1	8	0	2025-06-16 12:15:07	2025-06-16 12:15:07
2236	55	1	19	0	2025-06-16 12:15:08	2025-06-16 12:15:08
2237	55	3	1	2	2025-06-16 12:15:33	2025-06-16 12:15:33
2238	55	3	2	3	2025-06-16 12:15:33	2025-06-16 12:15:33
2239	55	3	3	0	2025-06-16 12:15:33	2025-06-16 12:15:33
2240	55	3	4	0	2025-06-16 12:15:33	2025-06-16 12:15:33
2241	55	3	5	0	2025-06-16 12:15:33	2025-06-16 12:15:33
2242	55	3	6	0	2025-06-16 12:15:33	2025-06-16 12:15:33
2243	55	3	8	0	2025-06-16 12:15:33	2025-06-16 12:15:33
2244	55	3	19	2	2025-06-16 12:15:34	2025-06-16 12:15:34
2245	56	7	1	7	2025-08-06 10:43:20	2025-08-06 10:43:20
2246	56	7	2	7	2025-08-06 10:43:22	2025-08-06 10:43:22
2247	56	7	3	0	2025-08-06 10:43:22	2025-08-06 10:43:22
2248	56	7	4	0	2025-08-06 10:43:23	2025-08-06 10:43:23
2249	56	7	5	0	2025-08-06 10:43:23	2025-08-06 10:43:23
2250	56	7	6	0	2025-08-06 10:43:23	2025-08-06 10:43:23
2251	56	7	8	0	2025-08-06 10:43:24	2025-08-06 10:43:24
2252	56	7	19	0	2025-08-06 10:43:24	2025-08-06 10:43:24
2253	56	12	1	5	2025-08-06 10:43:51	2025-08-06 10:43:51
2254	56	12	2	7	2025-08-06 10:43:52	2025-08-06 10:43:52
2255	56	12	3	0	2025-08-06 10:43:52	2025-08-06 10:43:52
2256	56	12	4	0	2025-08-06 10:43:52	2025-08-06 10:43:52
2257	56	12	5	0	2025-08-06 10:43:53	2025-08-06 10:43:53
2258	56	12	6	0	2025-08-06 10:43:53	2025-08-06 10:43:53
2259	56	12	8	0	2025-08-06 10:43:54	2025-08-06 10:43:54
2260	56	12	19	0	2025-08-06 10:43:54	2025-08-06 10:43:54
2261	56	13	1	8	2025-08-06 10:44:09	2025-08-06 10:44:09
2262	56	13	2	2	2025-08-06 10:44:09	2025-08-06 10:44:09
2263	56	13	3	0	2025-08-06 10:44:09	2025-08-06 10:44:09
2264	56	13	4	0	2025-08-06 10:44:10	2025-08-06 10:44:10
2265	56	13	5	0	2025-08-06 10:44:10	2025-08-06 10:44:10
2266	56	13	6	0	2025-08-06 10:44:10	2025-08-06 10:44:10
2267	56	13	8	0	2025-08-06 10:44:10	2025-08-06 10:44:10
2268	56	13	19	0	2025-08-06 10:44:10	2025-08-06 10:44:10
2269	56	2	1	3	2025-08-06 10:44:26	2025-08-06 10:44:26
2270	56	2	2	5	2025-08-06 10:44:26	2025-08-06 10:44:26
2271	56	2	3	0	2025-08-06 10:44:27	2025-08-06 10:44:27
2272	56	2	4	0	2025-08-06 10:44:27	2025-08-06 10:44:27
2273	56	2	5	0	2025-08-06 10:44:27	2025-08-06 10:44:27
2274	56	2	6	0	2025-08-06 10:44:27	2025-08-06 10:44:27
2275	56	2	8	0	2025-08-06 10:44:28	2025-08-06 10:44:28
2276	56	2	19	0	2025-08-06 10:44:28	2025-08-06 10:44:28
2277	56	1	1	7	2025-08-06 10:44:42	2025-08-06 10:44:42
2278	56	1	2	6	2025-08-06 10:44:42	2025-08-06 10:44:42
2279	56	1	3	0	2025-08-06 10:44:43	2025-08-06 10:44:43
2280	56	1	4	0	2025-08-06 10:44:43	2025-08-06 10:44:43
2281	56	1	5	0	2025-08-06 10:44:44	2025-08-06 10:44:44
2282	56	1	6	0	2025-08-06 10:44:44	2025-08-06 10:44:44
2283	56	1	8	0	2025-08-06 10:44:45	2025-08-06 10:44:45
2284	56	1	19	0	2025-08-06 10:44:45	2025-08-06 10:44:45
2285	57	10	1	1	2025-09-05 15:13:59	2025-09-05 15:13:59
2286	57	10	2	3	2025-09-05 15:13:59	2025-09-05 15:13:59
2287	57	10	3	0	2025-09-05 15:13:59	2025-09-05 15:13:59
2288	57	10	4	0	2025-09-05 15:13:59	2025-09-05 15:13:59
2289	57	10	5	0	2025-09-05 15:13:59	2025-09-05 15:13:59
2290	57	10	6	0	2025-09-05 15:13:59	2025-09-05 15:13:59
2291	57	10	8	0	2025-09-05 15:13:59	2025-09-05 15:13:59
2292	57	10	19	0	2025-09-05 15:13:59	2025-09-05 15:13:59
2293	57	8	1	7	2025-09-05 15:14:38	2025-09-05 15:14:38
2294	57	8	2	8	2025-09-05 15:14:39	2025-09-05 15:14:39
2295	57	8	3	1	2025-09-05 15:14:40	2025-09-05 15:14:40
2296	57	8	4	0	2025-09-05 15:14:41	2025-09-05 15:14:41
2297	57	8	5	0	2025-09-05 15:14:41	2025-09-05 15:14:41
2298	57	8	6	0	2025-09-05 15:14:41	2025-09-05 15:14:41
2299	57	8	8	0	2025-09-05 15:14:42	2025-09-05 15:14:42
2300	57	8	19	1	2025-09-05 15:14:42	2025-09-05 15:14:42
2301	57	9	1	5	2025-09-05 15:14:58	2025-09-05 15:14:58
2302	57	9	2	6	2025-09-05 15:14:58	2025-09-05 15:14:58
2303	57	9	3	0	2025-09-05 15:14:58	2025-09-05 15:14:58
2304	57	9	4	0	2025-09-05 15:14:58	2025-09-05 15:14:58
2305	57	9	5	0	2025-09-05 15:14:58	2025-09-05 15:14:58
2306	57	9	6	0	2025-09-05 15:14:59	2025-09-05 15:14:59
2307	57	9	8	0	2025-09-05 15:14:59	2025-09-05 15:14:59
2308	57	9	19	0	2025-09-05 15:15:00	2025-09-05 15:15:00
2309	57	11	1	3	2025-09-05 15:15:17	2025-09-05 15:15:17
2310	57	11	2	3	2025-09-05 15:15:18	2025-09-05 15:15:18
2311	57	11	3	1	2025-09-05 15:15:19	2025-09-05 15:15:19
2312	57	11	4	0	2025-09-05 15:15:20	2025-09-05 15:15:20
2313	57	11	5	0	2025-09-05 15:15:20	2025-09-05 15:15:20
2314	57	11	6	0	2025-09-05 15:15:20	2025-09-05 15:15:20
2315	57	11	8	0	2025-09-05 15:15:20	2025-09-05 15:15:20
2316	57	11	19	0	2025-09-05 15:15:21	2025-09-05 15:15:21
2317	57	2	1	4	2025-09-05 15:15:34	2025-09-05 15:15:34
2318	57	2	2	2	2025-09-05 15:15:35	2025-09-05 15:15:35
2319	57	2	3	0	2025-09-05 15:15:35	2025-09-05 15:15:35
2320	57	2	4	0	2025-09-05 15:15:35	2025-09-05 15:15:35
2321	57	2	5	0	2025-09-05 15:15:35	2025-09-05 15:15:35
2322	57	2	6	0	2025-09-05 15:15:35	2025-09-05 15:15:35
2323	57	2	8	0	2025-09-05 15:15:35	2025-09-05 15:15:35
2324	57	2	19	0	2025-09-05 15:15:35	2025-09-05 15:15:35
2325	57	1	1	4	2025-09-05 15:15:51	2025-09-05 15:15:51
2326	57	1	2	6	2025-09-05 15:15:51	2025-09-05 15:15:51
2327	57	1	3	1	2025-09-05 15:15:52	2025-09-05 15:15:52
2328	57	1	4	0	2025-09-05 15:15:52	2025-09-05 15:15:52
2329	57	1	5	0	2025-09-05 15:15:52	2025-09-05 15:15:52
2330	57	1	6	0	2025-09-05 15:15:52	2025-09-05 15:15:52
2331	57	1	8	0	2025-09-05 15:15:52	2025-09-05 15:15:52
2332	57	1	19	0	2025-09-05 15:15:52	2025-09-05 15:15:52
2333	57	7	1	5	2025-09-05 15:16:08	2025-09-05 15:16:08
2334	57	7	2	7	2025-09-05 15:16:09	2025-09-05 15:16:09
2335	57	7	3	0	2025-09-05 15:16:09	2025-09-05 15:16:09
2336	57	7	4	0	2025-09-05 15:16:09	2025-09-05 15:16:09
2337	57	7	5	0	2025-09-05 15:16:10	2025-09-05 15:16:10
2338	57	7	6	0	2025-09-05 15:16:10	2025-09-05 15:16:10
2339	57	7	8	0	2025-09-05 15:16:10	2025-09-05 15:16:10
2340	57	7	19	0	2025-09-05 15:16:10	2025-09-05 15:16:10
2341	58	10	1	2	2025-09-11 12:56:39	2025-09-11 12:56:39
2342	58	10	2	5	2025-09-11 12:56:39	2025-09-11 12:56:39
2343	58	10	3	0	2025-09-11 12:56:40	2025-09-11 12:56:40
2344	58	10	4	0	2025-09-11 12:56:40	2025-09-11 12:56:40
2345	58	10	5	0	2025-09-11 12:56:40	2025-09-11 12:56:40
2346	58	10	6	0	2025-09-11 12:56:40	2025-09-11 12:56:40
2347	58	10	8	0	2025-09-11 12:56:41	2025-09-11 12:56:41
2348	58	10	19	0	2025-09-11 12:56:41	2025-09-11 12:56:41
2349	58	8	1	12	2025-09-11 12:57:10	2025-09-11 12:57:10
2350	58	8	2	1	2025-09-11 12:57:11	2025-09-11 12:57:11
2351	58	8	3	4	2025-09-11 12:57:11	2025-09-11 12:57:11
2352	58	8	4	0	2025-09-11 12:57:11	2025-09-11 12:57:11
2353	58	8	5	0	2025-09-11 12:57:12	2025-09-11 12:57:12
2354	58	8	6	0	2025-09-11 12:57:12	2025-09-11 12:57:12
2355	58	8	8	0	2025-09-11 12:57:13	2025-09-11 12:57:13
2356	58	8	19	1	2025-09-11 12:57:13	2025-09-11 12:57:13
2357	58	4	1	11	2025-09-11 12:57:39	2025-09-11 12:57:39
2358	58	4	2	6	2025-09-11 12:57:40	2025-09-11 12:57:40
2359	58	4	3	0	2025-09-11 12:57:41	2025-09-11 12:57:41
2360	58	4	4	0	2025-09-11 12:57:41	2025-09-11 12:57:41
2361	58	4	5	0	2025-09-11 12:57:41	2025-09-11 12:57:41
2362	58	4	6	0	2025-09-11 12:57:41	2025-09-11 12:57:41
2363	58	4	8	0	2025-09-11 12:57:42	2025-09-11 12:57:42
2364	58	4	19	0	2025-09-11 12:57:42	2025-09-11 12:57:42
2365	58	9	1	12	2025-09-11 12:57:54	2025-09-11 12:57:54
2366	58	9	2	6	2025-09-11 12:57:55	2025-09-11 12:57:55
2367	58	9	3	0	2025-09-11 12:57:55	2025-09-11 12:57:55
2368	58	9	4	0	2025-09-11 12:57:56	2025-09-11 12:57:56
2369	58	9	5	0	2025-09-11 12:57:56	2025-09-11 12:57:56
2370	58	9	6	0	2025-09-11 12:57:56	2025-09-11 12:57:56
2371	58	9	8	0	2025-09-11 12:57:56	2025-09-11 12:57:56
2372	58	9	19	0	2025-09-11 12:57:57	2025-09-11 12:57:57
2373	58	11	1	3	2025-09-11 12:58:14	2025-09-11 12:58:14
2374	58	11	2	4	2025-09-11 12:58:15	2025-09-11 12:58:15
2375	58	11	3	0	2025-09-11 12:58:16	2025-09-11 12:58:16
2376	58	11	4	0	2025-09-11 12:58:16	2025-09-11 12:58:16
2377	58	11	5	0	2025-09-11 12:58:16	2025-09-11 12:58:16
2378	58	11	6	0	2025-09-11 12:58:16	2025-09-11 12:58:16
2379	58	11	8	0	2025-09-11 12:58:16	2025-09-11 12:58:16
2380	58	11	19	0	2025-09-11 12:58:17	2025-09-11 12:58:17
2381	58	7	1	11	2025-09-11 12:58:32	2025-09-11 12:58:32
2382	58	7	2	4	2025-09-11 12:58:32	2025-09-11 12:58:32
2383	58	7	3	0	2025-09-11 12:58:32	2025-09-11 12:58:32
2384	58	7	4	0	2025-09-11 12:58:32	2025-09-11 12:58:32
2385	58	7	5	0	2025-09-11 12:58:32	2025-09-11 12:58:32
2386	58	7	6	0	2025-09-11 12:58:32	2025-09-11 12:58:32
2387	58	7	8	0	2025-09-11 12:58:32	2025-09-11 12:58:32
2388	58	7	19	0	2025-09-11 12:58:32	2025-09-11 12:58:32
2389	58	1	1	6	2025-09-11 12:58:48	2025-09-11 12:58:48
2390	58	1	2	5	2025-09-11 12:58:49	2025-09-11 12:58:49
2391	58	1	3	1	2025-09-11 12:58:49	2025-09-11 12:58:49
2392	58	1	4	0	2025-09-11 12:58:50	2025-09-11 12:58:50
2393	58	1	5	0	2025-09-11 12:58:51	2025-09-11 12:58:51
2394	58	1	6	0	2025-09-11 12:58:51	2025-09-11 12:58:51
2395	58	1	8	0	2025-09-11 12:58:51	2025-09-11 12:58:51
2396	58	1	19	0	2025-09-11 12:58:52	2025-09-11 12:58:52
2397	58	3	1	5	2025-09-11 12:59:05	2025-09-11 12:59:05
2398	58	3	2	1	2025-09-11 12:59:06	2025-09-11 12:59:06
2399	58	3	3	0	2025-09-11 12:59:07	2025-09-11 12:59:07
2400	58	3	4	0	2025-09-11 12:59:08	2025-09-11 12:59:08
2401	58	3	5	0	2025-09-11 12:59:08	2025-09-11 12:59:08
2402	58	3	6	0	2025-09-11 12:59:09	2025-09-11 12:59:09
2403	58	3	8	0	2025-09-11 12:59:09	2025-09-11 12:59:09
2404	58	3	19	0	2025-09-11 12:59:09	2025-09-11 12:59:09
2405	60	10	1	3	2025-11-28 15:27:41	2025-11-28 15:27:41
2406	60	10	2	7	2025-11-28 15:27:41	2025-11-28 15:27:41
2407	60	10	3	0	2025-11-28 15:27:44	2025-11-28 15:27:44
2408	60	10	4	0	2025-11-28 15:27:44	2025-11-28 15:27:44
2409	60	10	5	0	2025-11-28 15:27:45	2025-11-28 15:27:45
2410	60	10	6	0	2025-11-28 15:27:45	2025-11-28 15:27:45
2411	60	10	8	0	2025-11-28 15:27:46	2025-11-28 15:27:46
2412	60	10	19	0	2025-11-28 15:27:46	2025-11-28 15:27:46
2413	60	8	1	2	2025-11-28 15:28:07	2025-11-28 15:28:07
2414	60	8	2	6	2025-11-28 15:28:08	2025-11-28 15:28:08
2415	60	8	3	0	2025-11-28 15:28:08	2025-11-28 15:28:08
2416	60	8	4	0	2025-11-28 15:28:09	2025-11-28 15:28:09
2417	60	8	5	0	2025-11-28 15:28:09	2025-11-28 15:28:09
2418	60	8	6	0	2025-11-28 15:28:09	2025-11-28 15:28:09
2419	60	8	8	0	2025-11-28 15:28:09	2025-11-28 15:28:09
2420	60	8	19	0	2025-11-28 15:28:10	2025-11-28 15:28:10
2421	60	9	1	8	2025-11-28 15:28:39	2025-11-28 15:28:39
2422	60	9	2	8	2025-11-28 15:28:39	2025-11-28 15:28:39
2423	60	9	3	0	2025-11-28 15:28:39	2025-11-28 15:28:39
2424	60	9	4	0	2025-11-28 15:28:39	2025-11-28 15:28:39
2425	60	9	5	1	2025-11-28 15:28:39	2025-11-28 15:28:39
2426	60	9	6	0	2025-11-28 15:28:39	2025-11-28 15:28:39
2427	60	9	8	0	2025-11-28 15:28:39	2025-11-28 15:28:39
2428	60	9	19	0	2025-11-28 15:28:40	2025-11-28 15:28:40
2429	60	11	1	7	2025-11-28 15:29:07	2025-11-28 15:29:07
2430	60	11	2	7	2025-11-28 15:29:07	2025-11-28 15:29:07
2431	60	11	3	0	2025-11-28 15:29:08	2025-11-28 15:29:08
2432	60	11	4	0	2025-11-28 15:29:08	2025-11-28 15:29:08
2433	60	11	5	0	2025-11-28 15:29:08	2025-11-28 15:29:08
2434	60	11	6	0	2025-11-28 15:29:09	2025-11-28 15:29:09
2435	60	11	8	0	2025-11-28 15:29:09	2025-11-28 15:29:09
2436	60	11	19	0	2025-11-28 15:29:09	2025-11-28 15:29:09
2437	60	1	1	9	2025-11-28 15:29:27	2025-11-28 15:29:27
2438	60	1	2	11	2025-11-28 15:29:28	2025-11-28 15:29:28
2439	60	1	3	0	2025-11-28 15:29:28	2025-11-28 15:29:28
2440	60	1	4	0	2025-11-28 15:29:29	2025-11-28 15:29:29
2441	60	1	5	0	2025-11-28 15:29:29	2025-11-28 15:29:29
2442	60	1	6	0	2025-11-28 15:29:29	2025-11-28 15:29:29
2443	60	1	8	0	2025-11-28 15:29:29	2025-11-28 15:29:29
2444	60	1	19	0	2025-11-28 15:29:30	2025-11-28 15:29:30
2445	61	10	1	9	2026-01-23 15:35:43	2026-01-23 15:35:43
2446	61	10	2	4	2026-01-23 15:35:44	2026-01-23 15:35:44
2447	61	10	3	0	2026-01-23 15:35:45	2026-01-23 15:35:45
2448	61	10	4	0	2026-01-23 15:35:45	2026-01-23 15:35:45
2449	61	10	5	0	2026-01-23 15:35:45	2026-01-23 15:35:45
2450	61	10	6	0	2026-01-23 15:35:45	2026-01-23 15:35:45
2451	61	10	8	0	2026-01-23 15:35:46	2026-01-23 15:35:46
2452	61	10	19	0	2026-01-23 15:35:46	2026-01-23 15:35:46
2453	61	8	1	7	2026-01-23 15:36:09	2026-01-23 15:36:09
2454	61	8	2	3	2026-01-23 15:36:09	2026-01-23 15:36:09
2455	61	8	3	0	2026-01-23 15:36:09	2026-01-23 15:36:09
2456	61	8	4	0	2026-01-23 15:36:09	2026-01-23 15:36:09
2457	61	8	5	0	2026-01-23 15:36:09	2026-01-23 15:36:09
2458	61	8	6	0	2026-01-23 15:36:09	2026-01-23 15:36:09
2459	61	8	8	0	2026-01-23 15:36:09	2026-01-23 15:36:09
2460	61	8	19	0	2026-01-23 15:36:09	2026-01-23 15:36:09
2461	61	4	1	15	2026-01-23 15:36:24	2026-01-23 15:36:24
2462	61	4	2	2	2026-01-23 15:36:24	2026-01-23 15:36:24
2463	61	4	3	0	2026-01-23 15:36:25	2026-01-23 15:36:25
2464	61	4	4	0	2026-01-23 15:36:26	2026-01-23 15:36:26
2465	61	4	5	0	2026-01-23 15:36:26	2026-01-23 15:36:26
2466	61	4	6	0	2026-01-23 15:36:26	2026-01-23 15:36:26
2467	61	4	8	0	2026-01-23 15:36:27	2026-01-23 15:36:27
2468	61	4	19	0	2026-01-23 15:36:27	2026-01-23 15:36:27
2469	61	9	1	8	2026-01-23 15:36:42	2026-01-23 15:36:42
2470	61	9	2	4	2026-01-23 15:36:43	2026-01-23 15:36:43
2471	61	9	3	0	2026-01-23 15:36:44	2026-01-23 15:36:44
2472	61	9	4	0	2026-01-23 15:36:44	2026-01-23 15:36:44
2473	61	9	5	0	2026-01-23 15:36:44	2026-01-23 15:36:44
2474	61	9	6	0	2026-01-23 15:36:44	2026-01-23 15:36:44
2475	61	9	8	0	2026-01-23 15:36:44	2026-01-23 15:36:44
2476	61	9	19	0	2026-01-23 15:36:44	2026-01-23 15:36:44
2477	61	2	1	12	2026-01-23 15:37:05	2026-01-23 15:37:05
2478	61	2	2	3	2026-01-23 15:37:05	2026-01-23 15:37:05
2479	61	2	3	0	2026-01-23 15:37:06	2026-01-23 15:37:06
2480	61	2	4	0	2026-01-23 15:37:06	2026-01-23 15:37:06
2481	61	2	5	0	2026-01-23 15:37:06	2026-01-23 15:37:06
2482	61	2	6	0	2026-01-23 15:37:06	2026-01-23 15:37:06
2483	61	2	8	0	2026-01-23 15:37:06	2026-01-23 15:37:06
2484	61	2	19	0	2026-01-23 15:37:06	2026-01-23 15:37:06
2485	61	7	1	7	2026-01-23 15:37:24	2026-01-23 15:37:24
2486	61	7	2	6	2026-01-23 15:37:25	2026-01-23 15:37:25
2487	61	7	3	0	2026-01-23 15:37:26	2026-01-23 15:37:26
2488	61	7	4	0	2026-01-23 15:37:26	2026-01-23 15:37:26
2489	61	7	5	0	2026-01-23 15:37:26	2026-01-23 15:37:26
2490	61	7	6	0	2026-01-23 15:37:26	2026-01-23 15:37:26
2491	61	7	8	0	2026-01-23 15:37:27	2026-01-23 15:37:27
2492	61	7	19	0	2026-01-23 15:37:27	2026-01-23 15:37:27
2493	61	1	1	8	2026-01-23 15:37:47	2026-01-23 15:37:47
2494	61	1	2	6	2026-01-23 15:37:48	2026-01-23 15:37:48
2495	61	1	3	2	2026-01-23 15:37:48	2026-01-23 15:37:48
2496	61	1	4	0	2026-01-23 15:37:49	2026-01-23 15:37:49
2497	61	1	5	0	2026-01-23 15:37:49	2026-01-23 15:37:49
2498	61	1	6	0	2026-01-23 15:37:50	2026-01-23 15:37:50
2499	61	1	8	0	2026-01-23 15:37:50	2026-01-23 15:37:50
2500	61	1	19	0	2026-01-23 15:37:50	2026-01-23 15:37:50
2501	61	3	1	7	2026-01-23 15:38:00	2026-01-23 15:38:00
2502	61	3	2	1	2026-01-23 15:38:01	2026-01-23 15:38:01
2503	61	3	3	0	2026-01-23 15:38:01	2026-01-23 15:38:01
2504	61	3	4	0	2026-01-23 15:38:01	2026-01-23 15:38:01
2505	61	3	5	0	2026-01-23 15:38:01	2026-01-23 15:38:01
2506	61	3	6	0	2026-01-23 15:38:02	2026-01-23 15:38:02
2507	61	3	8	0	2026-01-23 15:38:02	2026-01-23 15:38:02
2508	61	3	19	0	2026-01-23 15:38:02	2026-01-23 15:38:02
\.


--
-- Data for Name: fee_type_version_matchday; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.fee_type_version_matchday (fee_type_version_id, matchday_id) FROM stdin;
3	16
5	16
4	16
6	16
2	16
19	16
1	16
8	16
3	17
5	17
4	17
6	17
2	17
19	17
1	17
8	17
3	18
5	18
4	18
6	18
2	18
19	18
1	18
8	18
3	19
5	19
4	19
6	19
2	19
19	19
1	19
8	19
3	20
5	20
4	20
6	20
2	20
19	20
1	20
8	20
3	21
5	21
4	21
6	21
2	21
19	21
1	21
8	21
3	22
5	22
4	22
6	22
2	22
19	22
1	22
8	22
3	23
5	23
4	23
6	23
2	23
19	23
1	23
8	23
3	24
5	24
4	24
6	24
2	24
19	24
1	24
8	24
4	31
3	31
19	31
1	31
5	31
2	31
6	31
8	31
35	34
34	34
29	34
30	34
31	34
32	34
33	34
35	35
34	35
29	35
30	35
31	35
32	35
33	35
35	36
34	36
29	36
30	36
31	36
32	36
33	36
35	38
34	38
29	38
30	38
31	38
32	38
33	38
35	39
34	39
29	39
30	39
31	39
32	39
33	39
35	40
34	40
29	40
30	40
31	40
32	40
33	40
35	41
34	41
29	41
30	41
31	41
32	41
33	41
35	42
34	42
29	42
30	42
31	42
32	42
33	42
35	43
34	43
29	43
30	43
31	43
32	43
33	43
35	44
34	44
29	44
30	44
31	44
32	44
33	44
35	45
34	45
29	45
30	45
31	45
32	45
33	45
35	46
34	46
29	46
30	46
31	46
32	46
33	46
36	47
35	47
34	47
29	47
30	47
31	47
32	47
33	47
36	48
35	48
34	48
29	48
30	48
31	48
32	48
33	48
46	50
47	50
52	50
49	50
50	50
51	50
48	50
45	50
4	53
3	53
19	53
1	53
5	53
2	53
6	53
8	53
4	54
3	54
19	54
1	54
5	54
2	54
6	54
8	54
4	55
3	55
19	55
1	55
5	55
2	55
6	55
8	55
4	56
3	56
19	56
1	56
5	56
2	56
6	56
8	56
4	57
3	57
19	57
1	57
5	57
2	57
6	57
8	57
4	58
3	58
19	58
1	58
5	58
2	58
6	58
8	58
4	59
3	59
19	59
1	59
5	59
2	59
6	59
8	59
4	60
3	60
19	60
1	60
5	60
2	60
6	60
8	60
4	61
3	61
19	61
1	61
5	61
2	61
6	61
8	61
\.


--
-- Data for Name: fee_type_versions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.fee_type_versions (id, fee_type_id, name, description, amount, created_at, updated_at) FROM stdin;
9	9	verlorenes Spiel	\N	100	2025-03-09 10:33:50	2025-03-09 10:33:50
10	10	Gosse	\N	30	2025-03-09 10:34:19	2025-03-09 10:34:19
11	9	verlorenes Spiel	\N	30	2025-03-09 10:34:28	2025-03-09 10:34:28
12	11	Puppe vergessen	\N	500	2025-03-09 10:35:19	2025-03-09 10:35:19
13	12	Klingel	\N	100	2025-03-09 10:35:44	2025-03-09 10:35:44
14	13	Fluchen	\N	100	2025-03-09 10:35:59	2025-03-09 10:35:59
15	14	Kugel	\N	100	2025-03-09 10:36:16	2025-03-09 10:36:16
16	15	Lustwurf	\N	100	2025-03-09 10:36:58	2025-03-09 10:36:58
29	10	Gosse	\N	0	2025-03-24 17:34:04	2025-03-24 17:34:04
30	9	verlorenes Spiel	\N	0	2025-03-24 17:34:18	2025-03-24 17:34:18
31	15	Lustwurf	\N	0	2025-03-24 17:35:26	2025-03-24 17:35:26
32	12	Klingel	\N	0	2025-03-24 17:36:38	2025-03-24 17:36:38
33	11	Puppe vergessen	\N	0	2025-03-24 17:38:34	2025-03-24 17:38:34
34	13	Fluchen	\N	0	2025-03-24 17:38:44	2025-03-24 17:38:44
35	14	Kugel	\N	0	2025-03-24 17:38:52	2025-03-24 17:38:52
36	25	sonstiges	\N	0	2025-04-25 12:13:38	2025-04-25 12:13:38
37	11	Puppe vergessen	\N	500	2025-04-25 12:36:54	2025-04-25 12:36:54
38	14	Kugel	\N	100	2025-04-25 12:37:53	2025-04-25 12:37:53
39	25	sonstige Strafen	\N	10	2025-04-25 12:38:05	2025-04-25 12:38:05
40	10	Gosse	\N	30	2025-04-25 12:38:34	2025-04-25 12:38:34
41	9	verlorenes Spiel	\N	30	2025-04-25 12:38:41	2025-04-25 12:38:41
42	12	Klingel	\N	30	2025-04-25 12:38:48	2025-04-25 12:38:48
43	15	Lustwurf	\N	30	2025-04-25 12:38:55	2025-04-25 12:38:55
44	13	Fluchen	\N	30	2025-04-25 12:39:03	2025-04-25 12:39:03
45	11	Puppe vergessen	\N	0	2025-04-25 12:54:09	2025-04-25 12:54:09
46	25	sonstige Strafen	\N	0	2025-04-25 12:54:15	2025-04-25 12:54:15
47	14	Kugel	\N	0	2025-04-25 12:54:20	2025-04-25 12:54:20
48	12	Klingel	\N	0	2025-04-25 12:54:24	2025-04-25 12:54:24
49	10	Gosse	\N	0	2025-04-25 12:54:27	2025-04-25 12:54:27
50	9	verlorenes Spiel	\N	0	2025-04-25 12:54:31	2025-04-25 12:54:31
51	15	Lustwurf	\N	0	2025-04-25 12:54:35	2025-04-25 12:54:35
52	13	Fluchen	\N	0	2025-04-25 12:54:39	2025-04-25 12:54:39
1	1	Verlorenes Spiel	\N	-20	2025-03-08 11:36:07	2025-03-08 11:36:07
2	2	Gosse	\N	-20	2025-03-08 11:36:19	2025-03-08 11:36:19
3	3	Klingel	\N	-20	2025-03-08 11:36:54	2025-03-08 11:36:54
4	4	Lustwurf	\N	-20	2025-03-08 11:37:09	2025-03-08 11:37:09
5	5	Fluchen	Schimpfwort auf der Bahn benutzt	-100	2025-03-08 11:37:33	2025-03-08 11:37:33
6	6	Kugel	Mangelnde Aufmerksamkeit	-100	2025-03-08 11:37:59	2025-03-08 11:37:59
7	7	Puppe vergessen	Sieger- oder Verlierer-Puppe vergessen	-500	2025-03-08 11:38:18	2025-03-08 11:38:18
8	8	Sonstiges	Geldspiele etc.	-100	2025-03-08 11:38:33	2025-03-08 11:38:33
17	7	Puppe vergessen	Sieger- oder Verlierer-Puppe vergessen	-700	2025-03-10 20:03:48	2025-03-10 20:03:48
18	7	Puppe vergessen	Sieger- oder Verlierer-Puppe vergessen	-100	2025-03-11 10:32:22	2025-03-11 10:32:22
19	7	Puppe vergessen	Sieger- oder Verlierer-Puppe vergessen	-500	2025-03-15 16:23:37	2025-03-15 16:23:37
\.


--
-- Data for Name: fee_types; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.fee_types (id, club_id, name, description, amount, "position", created_at, updated_at) FROM stdin;
11	2	Puppe vergessen	\N	0	6	2025-03-09 10:35:19	2025-04-25 12:54:09
25	2	sonstige Strafen	\N	0	8	2025-04-25 12:13:38	2025-04-25 12:54:15
14	2	Kugel	\N	0	7	2025-03-09 10:36:16	2025-04-25 12:54:20
12	2	Klingel	\N	0	4	2025-03-09 10:35:44	2025-04-25 12:54:24
10	2	Gosse	\N	0	1	2025-03-09 10:34:19	2025-04-25 12:54:27
9	2	verlorenes Spiel	\N	0	2	2025-03-09 10:33:50	2025-04-25 12:54:31
15	2	Lustwurf	\N	0	5	2025-03-09 10:36:58	2025-04-25 12:54:35
13	2	Fluchen	\N	0	3	2025-03-09 10:35:59	2025-04-25 12:54:39
1	1	Verlorenes Spiel	\N	-20	1	2025-03-08 11:36:07	2025-03-08 11:36:07
2	1	Gosse	\N	-20	2	2025-03-08 11:36:19	2025-03-08 11:36:19
3	1	Klingel	\N	-20	3	2025-03-08 11:36:54	2025-03-08 11:36:54
4	1	Lustwurf	\N	-20	4	2025-03-08 11:37:09	2025-03-08 11:37:09
5	1	Fluchen	Schimpfwort auf der Bahn benutzt	-100	5	2025-03-08 11:37:33	2025-03-08 11:37:33
6	1	Kugel	Mangelnde Aufmerksamkeit	-100	6	2025-03-08 11:37:59	2025-03-08 11:37:59
8	1	Sonstiges	Geldspiele etc.	-100	8	2025-03-08 11:38:33	2025-03-10 20:03:42
7	1	Puppe vergessen	Sieger- oder Verlierer-Puppe vergessen	-500	7	2025-03-08 11:38:18	2025-03-15 16:23:37
\.


--
-- Data for Name: job_batches; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.job_batches (id, name, total_jobs, pending_jobs, failed_jobs, failed_job_ids, options, cancelled_at, created_at, finished_at) FROM stdin;
\.


--
-- Data for Name: jobs; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.jobs (id, queue, payload, attempts, reserved_at, available_at, created_at) FROM stdin;
\.


--
-- Data for Name: matchday_player; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.matchday_player (matchday_id, player_id, created_at) FROM stdin;
16	7	2025-03-31 15:52:23
16	8	2025-03-31 15:52:23
16	4	2025-03-31 15:52:23
16	9	2025-03-31 15:52:23
16	2	2025-03-31 15:52:23
16	3	2025-03-31 15:52:23
17	10	2025-03-31 15:52:23
17	8	2025-03-31 15:52:23
17	12	2025-03-31 15:52:23
17	4	2025-03-31 15:52:23
17	9	2025-03-31 15:52:23
17	13	2025-03-31 15:52:23
17	2	2025-03-31 15:52:23
17	7	2025-03-31 15:52:23
17	1	2025-03-31 15:52:23
17	3	2025-03-31 15:52:23
18	3	2025-03-31 15:52:23
18	8	2025-03-31 15:52:23
18	4	2025-03-31 15:52:23
18	9	2025-03-31 15:52:23
18	11	2025-03-31 15:52:23
18	2	2025-03-31 15:52:23
18	1	2025-03-31 15:52:23
18	10	2025-03-31 15:52:23
19	10	2025-03-31 15:52:23
19	7	2025-03-31 15:52:23
19	11	2025-03-31 15:52:23
19	2	2025-03-31 15:52:23
19	1	2025-03-31 15:52:23
20	10	2025-03-31 15:52:23
20	8	2025-03-31 15:52:23
20	4	2025-03-31 15:52:23
20	11	2025-03-31 15:52:23
20	2	2025-03-31 15:52:23
20	3	2025-03-31 15:52:23
21	10	2025-03-31 15:52:23
21	11	2025-03-31 15:52:23
21	2	2025-03-31 15:52:23
21	1	2025-03-31 15:52:23
21	7	2025-03-31 15:52:23
22	10	2025-03-31 15:52:23
22	9	2025-03-31 15:52:23
22	14	2025-03-31 15:52:23
22	2	2025-03-31 15:52:23
22	1	2025-03-31 15:52:23
22	5	2025-03-31 15:52:23
23	4	2025-03-31 15:52:23
23	8	2025-03-31 15:52:23
23	3	2025-03-31 15:52:23
23	11	2025-03-31 15:52:23
23	10	2025-03-31 15:52:23
23	13	2025-03-31 15:52:23
23	9	2025-03-31 15:52:23
23	7	2025-03-31 15:52:23
23	12	2025-03-31 15:52:23
24	3	2025-03-31 15:52:23
24	1	2025-03-31 15:52:23
24	2	2025-03-31 15:52:23
24	4	2025-03-31 15:52:23
24	7	2025-03-31 15:52:23
24	8	2025-03-31 15:52:23
24	10	2025-03-31 15:52:23
24	9	2025-03-31 15:52:23
31	4	2025-03-31 15:52:23
31	8	2025-03-31 15:52:23
31	12	2025-03-31 15:52:23
31	9	2025-03-31 15:52:23
31	2	2025-03-31 15:52:23
31	1	2025-03-31 15:52:23
31	11	2025-03-31 15:52:23
31	13	2025-03-31 15:52:23
31	3	2025-03-31 15:52:23
31	10	2025-03-31 15:52:23
34	30	2025-03-31 15:52:23
34	36	2025-03-31 15:52:23
34	32	2025-03-31 15:52:23
34	33	2025-03-31 15:52:23
34	35	2025-03-31 15:52:23
34	34	2025-03-31 15:52:23
34	37	2025-03-31 15:52:23
35	30	2025-03-31 15:52:23
35	36	2025-03-31 15:52:23
35	32	2025-03-31 15:52:23
35	34	2025-03-31 15:52:23
35	35	2025-03-31 15:52:23
35	33	2025-03-31 15:52:23
35	37	2025-03-31 15:52:23
35	31	2025-03-31 15:52:23
36	36	2025-03-31 15:52:23
36	37	2025-03-31 15:52:23
36	31	2025-03-31 15:52:23
36	33	2025-03-31 15:52:23
38	30	2025-04-25 11:57:30
38	31	2025-04-25 11:57:41
38	32	2025-04-25 11:57:48
38	33	2025-04-25 11:57:53
38	35	2025-04-25 11:57:56
38	34	2025-04-25 11:58:00
38	37	2025-04-25 11:58:05
39	30	2025-04-25 11:58:32
39	36	2025-04-25 11:58:34
39	33	2025-04-25 11:58:37
39	32	2025-04-25 11:58:39
39	34	2025-04-25 11:58:41
39	35	2025-04-25 11:58:45
39	31	2025-04-25 11:58:50
40	30	2025-04-25 11:59:11
40	36	2025-04-25 11:59:21
40	33	2025-04-25 11:59:25
40	35	2025-04-25 11:59:29
40	37	2025-04-25 11:59:33
40	31	2025-04-25 11:59:36
41	30	2025-04-25 11:59:53
41	31	2025-04-25 12:00:00
41	32	2025-04-25 12:00:05
41	33	2025-04-25 12:00:10
41	34	2025-04-25 12:00:16
42	34	2025-04-25 12:00:47
42	32	2025-04-25 12:00:54
42	37	2025-04-25 12:01:02
42	35	2025-04-25 12:01:08
43	30	2025-04-25 12:01:42
43	31	2025-04-25 12:01:47
43	32	2025-04-25 12:01:51
43	33	2025-04-25 12:01:57
43	36	2025-04-25 12:02:06
43	34	2025-04-25 12:02:07
43	35	2025-04-25 12:02:09
43	37	2025-04-25 12:02:12
44	35	2025-04-25 12:02:38
44	31	2025-04-25 12:02:42
44	32	2025-04-25 12:02:46
44	36	2025-04-25 12:02:51
44	37	2025-04-25 12:02:55
45	36	2025-04-25 12:03:25
45	32	2025-04-25 12:03:28
45	33	2025-04-25 12:03:30
45	34	2025-04-25 12:03:32
45	37	2025-04-25 12:03:34
45	35	2025-04-25 12:03:37
45	31	2025-04-25 12:03:39
46	30	2025-04-25 12:04:12
46	36	2025-04-25 12:04:15
46	32	2025-04-25 12:04:17
46	33	2025-04-25 12:04:18
46	35	2025-04-25 12:04:20
46	34	2025-04-25 12:04:22
46	37	2025-04-25 12:04:24
46	31	2025-04-25 12:04:25
47	30	2025-04-25 12:19:09
47	31	2025-04-25 12:19:14
47	32	2025-04-25 12:19:20
47	33	2025-04-25 12:19:25
47	34	2025-04-25 12:19:32
47	35	2025-04-25 12:19:37
47	36	2025-04-25 12:19:42
48	30	2025-04-25 12:25:06
48	31	2025-04-25 12:25:11
48	32	2025-04-25 12:25:15
48	33	2025-04-25 12:25:20
48	34	2025-04-25 12:25:24
50	35	2025-04-25 12:55:00
50	31	2025-04-25 12:55:05
50	32	2025-04-25 12:55:08
50	36	2025-04-25 12:55:13
50	37	2025-04-25 12:55:17
53	10	2025-06-13 15:34:14
53	8	2025-06-13 15:34:20
53	4	2025-06-13 15:34:29
53	9	2025-06-13 15:34:38
53	2	2025-06-13 15:34:42
53	7	2025-06-13 15:34:50
53	1	2025-06-13 15:34:54
53	3	2025-06-13 15:34:57
54	10	2025-06-13 15:39:44
54	8	2025-06-13 15:39:54
54	4	2025-06-13 15:39:58
54	9	2025-06-13 15:40:03
54	11	2025-06-13 15:40:08
54	2	2025-06-13 15:40:12
54	7	2025-06-13 15:40:23
54	1	2025-06-13 15:40:28
54	3	2025-06-13 15:40:32
55	10	2025-06-16 12:11:02
55	8	2025-06-16 12:11:10
55	4	2025-06-16 12:11:17
55	9	2025-06-16 12:11:26
55	11	2025-06-16 12:11:33
55	2	2025-06-16 12:11:43
55	1	2025-06-16 12:11:50
55	3	2025-06-16 12:11:58
56	7	2025-08-06 10:42:04
56	12	2025-08-06 10:42:11
56	13	2025-08-06 10:42:20
56	2	2025-08-06 10:42:31
56	1	2025-08-06 10:42:38
57	10	2025-09-05 15:12:53
57	8	2025-09-05 15:13:02
57	9	2025-09-05 15:13:09
57	11	2025-09-05 15:13:17
57	2	2025-09-05 15:13:22
57	1	2025-09-05 15:13:29
57	7	2025-09-05 15:13:35
58	10	2025-09-11 12:54:58
58	8	2025-09-11 12:55:06
58	4	2025-09-11 12:55:15
58	9	2025-09-11 12:55:25
58	11	2025-09-11 12:55:33
58	7	2025-09-11 12:55:39
58	1	2025-09-11 12:55:47
58	3	2025-09-11 12:55:54
60	10	2025-11-28 15:26:20
60	8	2025-11-28 15:26:33
60	9	2025-11-28 15:26:41
60	11	2025-11-28 15:26:53
60	1	2025-11-28 15:27:02
61	10	2026-01-23 15:34:05
61	8	2026-01-23 15:34:13
61	4	2026-01-23 15:34:21
61	9	2026-01-23 15:34:30
61	2	2026-01-23 15:34:40
61	7	2026-01-23 15:34:48
61	1	2026-01-23 15:34:56
61	3	2026-01-23 15:35:03
\.


--
-- Data for Name: matchdays; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.matchdays (id, club_id, date, notes, is_calculated, created_at, updated_at) FROM stdin;
16	1	2024-11-29	\N	f	2025-03-17 13:59:42	2025-03-17 13:59:42
17	1	2025-01-24	\N	f	2025-03-17 14:12:31	2025-03-17 14:12:31
18	1	2024-09-06	\N	f	2025-03-18 07:05:24	2025-03-18 07:05:24
19	1	2024-07-12	\N	f	2025-03-18 07:09:20	2025-03-18 07:09:20
20	1	2024-06-16	\N	f	2025-03-18 07:11:53	2025-03-18 07:11:53
21	1	2024-04-19	\N	f	2025-03-18 07:15:42	2025-03-18 07:15:42
22	1	2024-03-22	\N	f	2025-03-18 07:18:16	2025-03-18 07:18:16
23	1	2024-10-04	\N	f	2025-03-18 20:18:39	2025-03-18 20:18:39
24	1	2024-12-27	\N	f	2025-03-18 20:24:11	2025-03-18 20:24:11
31	1	2025-02-21	\N	f	2025-03-21 16:38:41	2025-03-21 16:38:41
34	2	2024-01-05	\N	f	2025-03-25 12:34:45	2025-03-25 12:34:45
35	2	2024-03-01	\N	f	2025-03-27 15:25:33	2025-03-27 15:25:33
36	2	2024-05-24	\N	f	2025-03-27 15:26:39	2025-03-27 15:27:12
38	2	2024-06-21	\N	f	2025-04-25 11:57:22	2025-04-25 11:57:22
39	2	0024-07-19	\N	f	2025-04-25 11:58:24	2025-04-25 11:58:24
40	2	2024-09-13	\N	f	2025-04-25 11:59:03	2025-04-25 11:59:03
41	2	0024-10-11	\N	f	2025-04-25 11:59:46	2025-04-25 11:59:46
42	2	0025-01-31	\N	f	2025-04-25 12:00:36	2025-04-25 12:00:36
43	2	0025-02-28	\N	f	2025-04-25 12:01:22	2025-04-25 12:01:22
44	2	0024-11-08	\N	f	2025-04-25 12:02:25	2025-04-25 12:02:25
45	2	0025-01-03	\N	f	2025-04-25 12:03:13	2025-04-25 12:03:13
46	2	0025-03-28	\N	f	2025-04-25 12:04:01	2025-04-25 12:04:01
47	2	2024-07-19	\N	f	2025-04-25 12:19:02	2025-04-25 12:19:02
48	2	2024-10-11	\N	f	2025-04-25 12:24:57	2025-04-25 12:24:57
50	2	2024-11-08	\N	f	2025-04-25 12:54:54	2025-04-25 12:54:54
53	1	2025-03-21	\N	f	2025-06-13 15:33:59	2025-06-13 15:33:59
54	1	2025-04-18	\N	f	2025-06-13 15:39:32	2025-06-13 15:39:32
55	1	2025-06-13	\N	f	2025-06-16 12:10:44	2025-06-16 12:10:44
56	1	2025-07-11	\N	f	2025-08-06 10:41:50	2025-08-06 10:41:50
57	1	2025-08-08	\N	f	2025-09-05 15:12:35	2025-09-05 15:12:35
58	1	2025-09-05	\N	f	2025-09-11 12:46:24	2025-09-11 12:46:24
59	1	2025-10-03	\N	f	2025-10-31 16:51:49	2025-10-31 16:51:49
60	1	2025-10-31	\N	f	2025-11-28 15:25:56	2025-11-28 15:25:56
61	1	2025-11-28	\N	f	2026-01-23 15:33:46	2026-01-23 15:33:46
\.


--
-- Data for Name: migrations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.migrations (id, migration, batch) FROM stdin;
1	0001_01_01_000000_create_users_table	1
2	0001_01_01_000001_create_cache_table	1
3	0001_01_01_000002_create_jobs_table	1
4	2025_02_26_073712_create_bouncer_tables	1
5	2025_02_26_080303_create_clubs_table	1
6	2025_02_27_080218_create_players_table	1
7	2025_02_27_080231_create_fee_types_table	1
8	2025_02_27_080238_create_fee_type_versions_table	1
9	2025_02_27_080348_create_matchdays_table	1
10	2025_02_27_080352_create_fee_entries_table	1
11	2025_02_27_080407_create_club_settings_table	1
12	2025_02_27_080417_create_transactions_table	1
13	2025_02_27_080431_create_competition_types_table	1
14	2025_02_27_080439_create_competition_entries_table	1
15	2025_02_27_080515_player_invitations_table	1
16	2025_03_08_084256_create_matchday_player_table	1
17	2025_03_08_094438_create_fee_type_version_matchday_table	1
18	2025_03_16_085741_create_dashboard_layouts_table	2
19	2025_03_29_132144_create_personal_access_tokens_table	3
20	2025_03_31_122302_alter_matchday_player_table	4
21	2025_04_09_064106_rename_bouncer_tables	5
22	2025_04_09_064358_create_permission_tables	5
23	2025_04_09_064609_alter_roles_table_add_base_fee	5
24	2025_04_09_065355_migrate_old_roles	5
25	2025_04_09_072652_alter_players_table_add_role_id	5
26	2025_04_09_074133_migrate_player_roles	5
27	2025_04_11_123043_drop_bouncer_tables	6
\.


--
-- Data for Name: model_has_permissions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.model_has_permissions (permission_id, model_type, model_id, club_id) FROM stdin;
\.


--
-- Data for Name: model_has_roles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.model_has_roles (role_id, model_type, model_id, club_id) FROM stdin;
2	App\\Models\\Player	1	1
4	App\\Models\\Player	35	2
3	App\\Models\\Player	2	1
2	App\\Models\\Player	5	1
2	App\\Models\\Player	7	1
2	App\\Models\\Player	8	1
3	App\\Models\\Player	11	1
2	App\\Models\\Player	14	1
1	App\\Models\\Player	12	1
1	App\\Models\\Player	13	1
1	App\\Models\\Player	6	1
\.


--
-- Data for Name: password_reset_tokens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.password_reset_tokens (email, token, created_at) FROM stdin;
\.


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.permissions (id, name, guard_name, created_at, updated_at) FROM stdin;
1	list.Role	web	2025-04-11 12:01:47	2025-04-11 12:01:47
2	list.Player	web	2025-04-11 12:01:47	2025-04-11 12:01:47
3	list.Matchday	web	2025-04-11 12:01:47	2025-04-11 12:01:47
4	list.CompetitionType	web	2025-04-11 12:01:47	2025-04-11 12:01:47
5	list.Transaction	web	2025-04-11 12:01:47	2025-04-11 12:01:47
6	view.Club	web	2025-04-11 12:02:09	2025-04-11 12:02:09
7	view.Role	web	2025-04-11 12:02:09	2025-04-11 12:02:09
8	view.Player	web	2025-04-11 12:02:09	2025-04-11 12:02:09
9	view.Matchday	web	2025-04-11 12:02:09	2025-04-11 12:02:09
10	list.FeeType	web	2025-04-11 12:02:09	2025-04-11 12:02:09
11	view.FeeType	web	2025-04-11 12:02:09	2025-04-11 12:02:09
12	view.CompetitionType	web	2025-04-11 12:02:09	2025-04-11 12:02:09
13	view.Transaction	web	2025-04-11 12:02:09	2025-04-11 12:02:09
14	create.Player	web	2025-04-11 12:02:29	2025-04-11 12:02:29
15	update.Player	web	2025-04-11 12:02:29	2025-04-11 12:02:29
16	delete.Player	web	2025-04-11 12:02:29	2025-04-11 12:02:29
17	create.Matchday	web	2025-04-11 12:02:29	2025-04-11 12:02:29
18	update.Matchday	web	2025-04-11 12:02:29	2025-04-11 12:02:29
19	delete.Matchday	web	2025-04-11 12:02:29	2025-04-11 12:02:29
20	create.FeeType	web	2025-04-11 12:02:29	2025-04-11 12:02:29
21	update.FeeType	web	2025-04-11 12:02:29	2025-04-11 12:02:29
22	delete.FeeType	web	2025-04-11 12:02:29	2025-04-11 12:02:29
23	create.CompetitionType	web	2025-04-11 12:02:29	2025-04-11 12:02:29
24	update.CompetitionType	web	2025-04-11 12:02:29	2025-04-11 12:02:29
25	delete.CompetitionType	web	2025-04-11 12:02:29	2025-04-11 12:02:29
26	create.Transaction	web	2025-04-11 12:02:29	2025-04-11 12:02:29
27	update.Transaction	web	2025-04-11 12:02:29	2025-04-11 12:02:29
28	delete.Transaction	web	2025-04-11 12:02:29	2025-04-11 12:02:29
29	update.Club	web	2025-04-11 12:20:06	2025-04-11 12:20:06
30	delete.Club	web	2025-04-11 12:20:06	2025-04-11 12:20:06
31	create.Role	web	2025-04-11 12:20:06	2025-04-11 12:20:06
32	update.Role	web	2025-04-11 12:20:06	2025-04-11 12:20:06
33	delete.Role	web	2025-04-11 12:20:06	2025-04-11 12:20:06
34	list.Role	player	2025-06-13 15:12:53	2025-06-13 15:12:53
35	view.Role	player	2025-06-13 15:12:53	2025-06-13 15:12:53
36	list.Player	player	2025-06-13 15:12:53	2025-06-13 15:12:53
37	view.Player	player	2025-06-13 15:12:53	2025-06-13 15:12:53
38	create.Player	player	2025-06-13 15:12:53	2025-06-13 15:12:53
39	update.Player	player	2025-06-13 15:12:53	2025-06-13 15:12:53
40	delete.Player	player	2025-06-13 15:12:53	2025-06-13 15:12:53
41	list.Matchday	player	2025-06-13 15:12:53	2025-06-13 15:12:53
42	view.Matchday	player	2025-06-13 15:12:53	2025-06-13 15:12:53
43	create.Matchday	player	2025-06-13 15:12:53	2025-06-13 15:12:53
44	update.Matchday	player	2025-06-13 15:12:53	2025-06-13 15:12:53
45	delete.Matchday	player	2025-06-13 15:12:53	2025-06-13 15:12:53
46	list.CompetitionType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
47	view.CompetitionType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
48	create.CompetitionType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
49	update.CompetitionType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
50	delete.CompetitionType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
51	list.Transaction	player	2025-06-13 15:12:53	2025-06-13 15:12:53
52	view.Transaction	player	2025-06-13 15:12:53	2025-06-13 15:12:53
53	create.Transaction	player	2025-06-13 15:12:53	2025-06-13 15:12:53
54	update.Transaction	player	2025-06-13 15:12:53	2025-06-13 15:12:53
55	delete.Transaction	player	2025-06-13 15:12:53	2025-06-13 15:12:53
56	view.Club	player	2025-06-13 15:12:53	2025-06-13 15:12:53
57	list.FeeType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
58	view.FeeType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
59	create.FeeType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
60	update.FeeType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
61	delete.FeeType	player	2025-06-13 15:12:53	2025-06-13 15:12:53
\.


--
-- Data for Name: personal_access_tokens; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.personal_access_tokens (id, tokenable_type, tokenable_id, name, token, abilities, last_used_at, expires_at, created_at, updated_at) FROM stdin;
\.


--
-- Data for Name: player_invitations; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.player_invitations (id, player_id, email, token, expires_at, created_at, updated_at) FROM stdin;
2	2	nicole@schnurbus.de	1gg0WaBvCS4xlyoWsuVpKxWfRJoIjTQcKqgBUdzDOYoz4vUHcDAnMbJ2sGDI	2025-03-15 11:42:40	2025-03-08 11:42:40	2025-03-08 11:42:40
5	35	pascal@schnurbus.de	6762P6dCi1zV5sMYwzfsb2zz3BZWw0ETV8CpJLVg85kUaxfvuKoB3hm3er0J	2025-03-31 17:31:33	2025-03-24 17:31:33	2025-03-24 17:31:33
\.


--
-- Data for Name: players; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.players (id, name, club_id, user_id, sex, balance, initial_balance, active, created_at, updated_at, role_id) FROM stdin;
6	Monique	1	\N	2	0	0	f	2025-03-08 11:40:19	2025-06-13 15:35:43	1
5	Daniel	1	\N	1	0	0	f	2025-03-08 11:40:10	2025-06-13 15:10:52	2
13	Stephi	1	\N	2	-200	0	t	2025-03-08 11:41:33	2025-08-06 12:07:08	1
30	Dacky	2	\N	1	-4390	-4390	t	2025-03-24 16:44:53	2025-04-25 12:34:34	5
11	Sarah	1	\N	2	-1000	0	t	2025-03-08 11:41:11	2026-01-23 15:33:46	3
10	Michael	1	2	1	-1260	0	t	2025-03-08 11:41:03	2026-01-23 15:35:45	2
8	Klausi	1	\N	1	-1200	0	t	2025-03-08 11:40:42	2026-01-23 15:36:09	2
9	Nicki	1	\N	2	-1240	0	t	2025-03-08 11:40:53	2026-01-23 15:36:44	2
2	Nicole	1	3	2	-1300	0	t	2025-03-08 11:39:21	2026-01-23 15:37:06	3
7	Dacky	1	\N	1	-1260	0	t	2025-03-08 11:40:32	2026-01-23 15:37:25	2
3	Sven	1	\N	1	-1160	0	t	2025-03-08 11:39:40	2026-01-23 15:38:01	2
14	Carina	1	\N	2	-200	0	f	2025-03-18 07:18:44	2025-08-06 12:04:57	2
12	Mathias	1	\N	1	-240	0	t	2025-03-08 11:41:20	2025-08-06 12:07:30	1
33	Mathias	2	\N	1	0	0	t	2025-03-24 16:45:53	2025-04-25 12:35:03	5
34	Micha	2	\N	1	0	0	t	2025-03-24 16:46:12	2025-04-25 12:35:28	4
4	Bine	1	\N	2	-1340	0	t	2025-03-08 11:39:56	2026-01-23 15:36:25	2
31	Thilo	2	\N	2	0	0	t	2025-03-24 16:45:10	2025-04-25 12:53:52	5
32	Klaus	2	\N	1	0	0	t	2025-03-24 16:45:28	2025-04-25 12:53:52	5
1	Pascal	1	1	1	-1320	0	t	2025-03-08 11:39:08	2026-01-23 15:37:49	2
36	Daniel	2	\N	1	0	0	t	2025-03-24 16:46:59	2025-04-25 12:53:52	5
37	Sven	2	\N	1	0	0	t	2025-03-24 16:47:13	2025-04-25 12:53:52	5
35	Pascal	2	1	1	0	0	t	2025-03-24 16:46:37	2025-04-25 13:46:34	4
\.


--
-- Data for Name: role_has_permissions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.role_has_permissions (permission_id, role_id) FROM stdin;
6	4
29	4
30	4
1	4
7	4
31	4
32	4
33	4
2	4
8	4
14	4
15	4
16	4
3	4
9	4
17	4
18	4
19	4
10	4
11	4
20	4
21	4
22	4
4	4
12	4
23	4
24	4
25	4
5	4
13	4
26	4
27	4
28	4
2	6
6	5
1	5
7	5
2	5
8	5
3	5
9	5
10	5
4	5
12	5
5	5
13	5
11	5
6	13
29	13
1	13
7	13
2	13
8	13
14	13
15	13
16	13
3	13
9	13
17	13
18	13
19	13
10	13
11	13
20	13
21	13
22	13
4	13
12	13
23	13
24	13
25	13
5	13
13	13
26	13
27	13
28	13
34	3
35	3
36	3
37	3
38	3
39	3
40	3
41	3
42	3
43	3
44	3
45	3
46	3
47	3
48	3
49	3
50	3
51	3
52	3
53	3
54	3
55	3
56	3
57	3
58	3
59	3
60	3
61	3
34	2
35	2
36	2
37	2
41	2
42	2
46	2
47	2
51	2
52	2
56	2
57	2
58	2
34	1
36	1
41	1
46	1
51	1
\.


--
-- Data for Name: roles; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.roles (id, club_id, name, guard_name, created_at, updated_at, is_base_fee_active) FROM stdin;
1	1	Gast	player	2025-04-11 12:00:16	2025-04-11 12:00:16	f
2	1	Mitglied	player	2025-04-11 12:00:16	2025-04-11 12:00:16	t
3	1	Kassenwart	player	2025-04-11 12:00:16	2025-04-11 12:00:16	t
4	2	Admin	player	2025-04-11 12:00:16	2025-04-11 12:00:16	t
5	2	Spieler	player	2025-04-11 12:00:16	2025-04-11 12:00:16	t
6	2	Gast	player	2025-04-11 12:00:16	2025-04-11 12:00:16	f
7	3	owner	player	2025-04-11 12:00:16	2025-04-11 12:00:16	f
8	3	Spieler	player	2025-04-11 12:00:16	2025-04-11 12:00:16	f
9	4	owner	player	2025-04-11 12:00:16	2025-04-11 12:00:16	f
10	5	Gast	player	2025-04-11 12:00:16	2025-04-11 12:00:16	f
11	5	Mitglied	player	2025-04-11 12:00:16	2025-04-11 12:00:16	t
12	5	Kassenwart	player	2025-04-11 12:00:16	2025-04-11 12:00:16	t
13	2	Kassenwart	player	2025-04-11 12:00:16	2025-04-11 12:00:16	t
\.


--
-- Data for Name: sessions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.sessions (id, user_id, ip_address, user_agent, payload, last_activity) FROM stdin;
iR5c5gmMlzbhpC00wcWI7XHunIwbenELNuPScucn	\N	34.165.80.145	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiaENwUDB4NmFLbnpBdk5yMXU2UWVWOFBkNzk5MTA2OVZJVzVjZGttVSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770083259
Xan2BNOkSZWzwBnI1mYDVBJtNJSeo19sgBUHirS6	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiaVpHN0VpQlF5RXFOT1N5RnhwcjNNU1Rjc3g2MWxzdGcyUk03NzB1SiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770083814
YflgmN45gmajdhIw9Vg0EVvwx9O477ErBPBNnp7o	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiQ29mdjdqQTFzVnZneG1Sd3pGTUNDVGx0SWphVlN5d3hXY21tMk9BQyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770083816
opglyrbViM1jB9MWzbdrflUATOMe3rZRYab9xJHQ	\N	108.61.119.153	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZFpvcW5SNk5EdWxVeEFDQ1IyQ0s4S3EyZ254d1VaQks4ODFvdXVrSSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770084163
MnRbgED89n0oWsp0NM6rRni2RUowgtbxfzESa3BT	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiazNUUXBYYVhmajZMdENIMzZvZ2ZTTVhibk1VUm13NkxrblVDODJybyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770084719
uOsGR3kjo2TEJuhoqNRugUxsMy4jexDKAtb7IRFm	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiRHROS2xKNGRoTkF2MkxRTGgxY0Z3cjloMUlKNHd3RnhuaU1MNDFKYiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770084721
0YbFFa3jVTf7Fm4bTYRBl6iKiGJk7E6bkAiUaaMC	\N	34.165.13.190	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoibXVCRWxLelhyeGw5ZExVcGhGY00zZ1pvdExUdFJlTldhQ2NnamJwTSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770085067
JUcf4SK9ollYwPHUImij82TFHZZgAGaEeprsaDqL	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZG1vb09pQk9kTngxUzg1TE0yTjY3aEUwdHQ4a0ZrcDR3V3BhSEJQaSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770085629
Lk7ZKBxP2nArMXDa25HZ15Pdu1ZZ2y5zMjQ0EvLI	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiTGduaE54YVdRTWNvbGVJNnBHN2VpeG9TTDFXNXRBckdmOW5EeWxrayI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770085631
SrvfhvPnjf3fdFOmTv0TKVlmHmn7JOezg21CapOM	\N	34.147.219.142	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoicmRjd1FYQ05hQWxOaFA1YVhTbzNQMEk2UWlzc1R3RE5CV2hiM3hqUiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770085971
xok9YFC4WW2DE7ZjlTd9iAqN7qt2RIcsvnOlPjRC	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZGM2N3g2U0FXQklGV1d5NFRWZU0xbjJ3enA5cjdhSzJpNUI1V3N6ZSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770086535
PEUdCurdx7iZCveEGCE7M3WAlV8VNZY2pjaofIZh	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiYk1sbEVJMER6dXZ6UHlDVVVRYlR5eVZXN3VDN1N4OTQ1VmtnOVdxTiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770086536
ghHEOvffo2HakzkTpgriRSMH3AKCZvQ0rvbQzxHS	\N	66.42.84.197	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiWnlBeUN3emFmUHhEWE1jYVhtekhHRVA1cjBMQUdQT3FqODJuREl2NSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770086876
dlx2OncVIKvwEz5F824kKyCYk0Wf863StaFlRiIQ	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoicjF2eXlmR0lDZGRHOWMyUzFIck80eHNPTzZCQ2c5RWFobnhpeGw3MCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770087440
MP5GZ8cuUOkzEL25h9O0Dk2hrF8PX71wjwUkWG82	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZmZqSDZWaU44TjVNeXFFWFplYmVzNUVZNTZYaG45TmRkbVNPVTFCQSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770087442
mFNeICgQ250TwBLqJ25B11UuT3Hu4xvZ9xqxnSbo	\N	15.161.88.159	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiRXNZbUNTS09Yb3RkTjZGN1BYb1FYMTJRVnJoYllrUEp4bFBNcWpWNyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770087780
Fy9y0xDPJpqShMyDDjmemSZpwnr8zkQUtQet45Dr	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiQllmU3V4TE1CenJaZGRyVFdjNGhsUUVrclBQYUtCbUU0RENvdXNCbyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770088346
wlUUuwgsvZIVELWtpyYmzT2lwck3MtenbSMuusK0	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiMmRsWUZJUTJDOUdrZ2NMSHpDZnJQRmdSSVpKbzQyaVV4UTVmZUFGQSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770088347
aSOVkmZj48eGSzOqqk19O5zEwvmRzRDOSlCvIBV1	\N	45.32.166.195	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiS3pLeTRnSzhLVkR6V2lqTVAzOVhFWTZUd3NPaFRGd0p6bmFDSHlpMCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770088684
kOQd1X4uT4tA86WxPpMxUzq2iQOtSRAXQpnu2UOe	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiY0FPSEFsVE5PaUVxMjYxNDhqY1MzSGcxbE1FUjBnaUxDRmxJZDg5WCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770089251
pe8dQTwdlDhv3THMYV8AVekexpzkHpZ3hFcyrpDk	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUGRDVXhwWnhpVUxLRlBudW5hb3NweFdmS01RY2x5cW1FbjVsRkZoZyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770089253
0ysgew4lkNRi75nakdvKhwhkjBmiig91JBDgkvTt	\N	45.63.61.213	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiWWR6M0VMbDRGZ3l6U2hTTlRhdmVvUDlGWXFWR1RrRkYyZmJVY3JiWCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770089589
w5aJbiD42CULl5anlNxuZZg8ePy5zXTYCR5B7zpP	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiY3RyQ0FSRjlNR1lzWXQ1ME03dlhsNVExZHA1bVlGYlZkdEh3YVdNbSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770090157
u0M1XrKS2U4Ky3unuvc5ifw32sVqwp4quqq83lAp	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiOTR3TDdjV2F4ZFJQV2gzeFBVUG9JR095Z0M4ZVVHcWRSbTVhZWpwWiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770090159
VcE5A0ZphwzLHE3PTrmTEbAIPDgqY2Hr7lZTY6jH	\N	68.183.39.102	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoidFhhZzQxTGhUYnp1NHM4TXV0em5Nc0haZWdIYjdYYW5EWklMaGVBdiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770090492
PPfY7UlCmFpn5DDmBHQaM037VTHIJqtQFfW0TrVn	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoidzVsQWZRNjBJeG56MWt2cXZLa1RlRG9UbDROOFMxb3g0MTZUZlk5bCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770091063
ApNukdhQjP6ODdyQi6l33KMvKayFX7qgEbR8Lq5W	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUzBsengyQW5jQ0dMSFBOREY3ek5uckVJWkZMWk53YkloTEdxU0JwZyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770091064
5GpDpWu1827QMhdqrAxDYBFTxir3WMlnUNHtEYwy	\N	35.197.195.213	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiMFVtSnhtM0V5SEcxTUVmRVNVdEhxcUNzcnliR2VINXdUc3FWVWRIciI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770091395
lSBOwyqZqAe8Tzmw82dsdHyFLv8ZyZU74pguPw48	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiOERIZUdmb053UHYwWHF0VDZTajJvZkljMkhOZEZtTDJad1pSN0FLWCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770091969
sJXQ8bgUfAiaG6fipJR4fdvoJMoG81ljHsZWsEQx	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiRnNSYjdac2FXSVpiZjVlZzdYRFpNZ3BNRHc0a0hlcElLRHRyMVcwTCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770091972
V6Pw8pkk8DUWgnVVVLgtt7pz3WzitZBCfOvD4OHa	\N	34.92.43.64	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiSjZnZU9RTUVlUDdFSWtmc05hSFJCMUZuRTNNbWxQY3pWWk42ajNvRiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770092300
WsckWKY2bsQrRIoKpDVNN2AtezkI8ednpxFoieHn	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiemt2RE9TTVNUWERkTmFYN0ZaWUVDclJ5U1I2aHBLSE9DWlhFclFEWiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770092889
uvvkNQqSpJYfNblKY2xGbpt7rljcnznOayyILPzq	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiRE5ieUk2Mmg2ZVRzS29tZk94T3pYbU5tV0ZzOFlYV1Q0WTBwcHFPVyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770092892
6omCygSeJgf7e60G4mJ7Ptr0hPoP1fzWGfdheI1W	\N	143.244.178.189	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiclAxNTdLVlV1dVowM1hwQjJ3TlZEbGJtOGxWTkhrb3BRWmM5WDRmWSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770093204
qlwbnSTmQHtgNo44WY535unFq5QdSt8b5naunvzp	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiRUdzenlaQzQxbFJoZkpIU09qZ2ZnbXJ6RHpNaTN2cjBGRXhvR2JMRiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770093801
AkSQlmjRZGyYM4iy9yXvkHV3HLbTKmewS14u8Z6n	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoidzFDMTF3d2Rlb1ByeUs0TVVsMVJPYnNSbW95d240QmpCNzVTTzlvUCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770093803
7MFL2ixk8iyIW0LzVY2XtYpKeHqYIwuh9dqRVDLF	\N	34.65.116.80	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiNldaVndpdkNPeVV4aGpyN21DVGFLRXk3c2dtelROU3UzTlZpVUtHbCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770094107
BEqolpVOeDHbUW9ESKTOQ0qdEEt7WRrlZT6TBADp	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoidXF4R2hHWFVSeE1jcVAzMDk4c3cyWnp4bm5QZUZ3bDN5U3NNbW0yTiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770094708
Jod3q0HgYnRTAQGb5GlUVkoo31S2VH8Evtuh0ALu	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiNDZiRGxvTTY1MnFjWUdtU1A5emVrNHVtU0xHUWdYTjlVb2RGSHhQNiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770094709
46YskrqfISRComO9ndYkvb8WBzsspIwjecXT9pPQ	\N	34.34.185.236	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiTDBGTDZmYzFNVWxXZ2ZCUEZDaGl6djdWQzBWOVFlWXV2eDdJTGJDTyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770095011
mng3trYOXih7wONhV4kBsAVGDu6XoHVBgSxG5MkU	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiSXFWamlrUnpxNVBVMFNSbDM2RlFDT3dlNm5rbTZrQ0ZxR1F1dlg1dCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770095611
TftouW6avztvlreQSwLahSqzB6ANLRYGLw5ZTF6V	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiNDNDY2lZWDRDWnhJSXRyemtLeUIzWExXRTZvc3J3cFlyRDZKVnVEeSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770095613
1NkMkvRXJWIYIl9ZAbUcVI3E0pvJmlHP97WE36as	\N	34.34.97.213	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoicEVEcnprcFhTTWNTOGZaYkhUUnJpNUZYNEFsZWYyYlE3bG1GWFh3UiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770095914
O0tVw4mE0rNgWbuRlZzF98XI005SJ8FkvD7K5B86	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoianJOdFdaSzhWMGlUbnN6dzdXNXJoTHNxRnp2V2Z1em9GR0FMTlJudyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770096522
vW3zR9fcTnRg6cWAe8HarZjagUbBy1Y4Ae8RWAAy	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiSXVlVm5WQVBiUkpINjZPNXVTVkV1QXQzdGRYa1dhaGo2NHRhamRGSyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770096526
2a0ru7G6FViAJb98oioAZttzyQnqciVGduMZKy3L	\N	35.197.195.213	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUnlSWkZvYjVXa0tObUkzVXdTMEpMbFViTHlzRndVeTNscUtYSnh1aCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770096818
pI5Vrnk1XEbx5nXtjgHSn21giOlYowR02JxeXrEA	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoieXJwV3hvTlIydWNDblM0MFFGVU9yd1dTWjFlM2liakdKREVGNEdqZiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770097432
5Gx9HC8AAExJs3heNhJVVQXCUfni0gC8j0UwUamL	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoib3JvbHNlUDlqSUJPOEFla2RCZTFoZmc0YjJWSzVyRXdNaUNQZVRxdSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770097434
mRcwbqM6J0o5zUFi9wyMSpm0ey1ng2lMc9fxmdDP	\N	15.160.68.179	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoieVRxSHhaN0l2RnVsb0EzR3Rkd01Lc2hacGd5OWIxWm1jOUhLWGt4RCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770097721
iHj0rOVj5OWc75sFv7WLIRW0nQIJcHkq7Wpw6W13	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUEI3bU1SbHQ4UDNTVGRaMjBqdmlDYWR2eGZBS1p6NWtTSGQ1RmRNViI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770098338
vF58C2bDfIrbZ7ULxUyAOLZvzSc7r9WyzMj2M5ji	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZ3Fpc0xiTDAyY005cEIzRlVuUml0WDZtR3psb2Faek5WUWVRTUNNViI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770098340
GjOYDx1LHcqbZqCPQGEcB617bvLhx2p8Es5ZBj4a	\N	165.22.125.189	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiWEd0aXZoRTdteUpZVGlscFlSMnR1UDY2QkpGV0xOTmQya2dUYkdUZSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770098623
NEWx90iKyyXks5o6AKHJTu1WBfkGC41F3VUPIalZ	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUXM4aDR1MzVzaW1CRHNnWGpJZmpINW12WFg2VHJIbndsd3hEa3g0UyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770099243
uzDo3x9GfXrQJpk555SGB4Ktmw56ht8G690lkPM7	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoicE04Z0JJMEZUUVhXMm53QjBraXNqYXNGdVFTOTl5c0NmOVpMaDBBYiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770099245
lAgktaQOVqRvUVulKKptfj4yWpTLrIy32n5LqTXH	\N	45.32.193.13	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiYTE3bWlsVmtwVGJEU3lISGRWdktvNGVZQmJ1ODZlU08zS05uRWs0NiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770099529
0wZYqCUpgApuP1zQFYAcaWLWZEfgBE3pWK7KosRw	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoibE83RHRtMWJOVmNKbmlYR0trZk1hZllwaUZ0S2NDajN6ZEZKdHhkOCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770100149
rZGK0eFDdFBpIPrIawhenhFVI09ecJOIKE8n41Xe	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiVmp1NEpZVW5KR1MyZ21McFlMTzhEanpSdVd1WDBKbjQ2MTF2OXk1dSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770100151
YPfT1lGiFpArFO3t7bgJQW60GMUgnq4A5KOgQ8el	\N	161.35.219.158	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiTW0yOXA0SnJ6U29nckNsOFI2dExHVjlZMW1YaHlyNGZ4anVUeHNQYSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770100435
lzIbuTa57UOE6pZmjtCY92gRReyJvLR1bvDv5V0l	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiMU5RMU0zVW9VdlNOTXB0OVBDY2JDU0FIY2dOajNYWTRDdVBaM1M3RyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770101067
8RsQaWOR7jH3YVrX5KhidydVxnkzkjQL1yayIFwV	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiTG5TaE5teWpORk83cWtPa1Y0RXF0eVZaUmJqNnV6Ym9hSk1vcWpTRyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770101069
Ro7TQt44EZLJ046wRgTsnN1lLaY34MKNlFxUk4bK	\N	206.189.178.14	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoib0plcjJsaDRJMjhldnNsRGZLVFpucGtDbnpwUUJsOVpRSndsQ296TCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770101340
ShTLQoxgckH5yQUbNJgEFUFta5vLO40aOBV1tkrE	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiRGVTb0tKaWkzU0VabHhrTGtKR1BQdFhwS0N3dkJva2JPTGY3OXBVbiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770101976
Y4uzKiEd68DMeZBmmHdzUcYgt7q3n4Pux5uXsE5K	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiVXM4MWw1R3NqNllkRzFhb2xuclZHaDNBUmRrU1RXTk1HbFFrbU1hViI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770101978
ekTQcFYqlZNX3cbPGqmeGTjPy6RU9F5lgqMPmMNE	\N	192.81.222.231	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiN3hHNFNzNlhieWduS0lLTHFxa1dNQjd1OU1EVnNNWjY2N2xvcUJBbCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770102244
CPaN6OeZQL0MAVoaq3KzvVNEE6awehTguKMDF2RH	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiVXZtZE9ISGFvamdpT3Y4b3NTbmFPRVBoOVpCVUE4YThTYmpmemZYTCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770102881
cCYP03S3nJVBZMXvAYw4w96PyFVCDoFXbxJQT4Xs	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiblFNcUtOSG11eG9jdTduTEVZQ3NuSVpGSFJoWlpEWEhjRjFtZ0h3NyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770102882
4LwK0EJ3ii3YR6AXXJM52yiwALFcBsR3lZCLpncV	\N	34.155.143.57	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUlZNaW5TZDI1ZDU5RTRMWjlYYTdpVWdiSElVdUVjSWZvaDlqbHBJcSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770103149
PsnQWlecsUoHx8DL9W9xTbvQZbw7zeT2bw5MFv3N	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUFQyZmF0cDduc01hb2NZR0JocUszQ0lxODRXSTJLMnNWQ2pMV3pYQSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770103785
rnQxtUnkMBKp4dh2NX7fXuQSxcZU5kXNrxQG6MD8	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiNXJLTmhoYjdXa1JmcTJmS1NHYktyalkwOHR0VGJ5QXBTSjdmS0gzeiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770103787
lBO7dtFxTS2micSflmIcT6VkBob9cJD8lGwXbkCj	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiTWpjeGN5Wk9RRFY2VXp5SmE1R2xzT0JxWWN0ZnlYdmxVdkN4SENucyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770103787
KFHGCsyx1c5f5nsFqGB0lEjj7wHqZ3kSkLLNSeGI	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoicGpNRzNUcEFrdTBEWkpicXdtemUyY2s1dEpndGdOUUVudW1IVFZKTiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770103789
WrfFsczc6c1CUbr38j50RsGmXyaZbR3GidZz82m3	\N	34.175.19.199	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiYnpOUVAxMlFnRjh5TEpGcmsyREpCOGNrOXRVSU9vVXJHODFmU2kxTCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770104051
diuwdUVPJAID0bzLkva8XV3G9RIqbz9ajbxOLUlG	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiMU1FN05tWTN2bXE2SE91d0Z4elY4VDl4anptSkNMTlpwN3R6amR6VCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770104697
0wQ6ig49LKa5YEDB7G9cRctHm1gD3ssbZOPQ86JJ	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiVkd1dTJXQTdXckZYdlVCREhqeHJsU2pHUEc5eGF6czZ6QnFJdHVrSCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770104699
eQMqbzy90Ac5SPPxSlm7RlXBKrpi8aeqb0a5e5Z1	\N	146.190.20.113	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiNlZnUkdndjZ2OWlxUTNFWFJVT0NaaTM4cXpJR255Nnd0TFl3OFRJMiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770104954
4DcSrUVGQ1ukkvczWSRJrsRxfe31O3PeY7mtCFRX	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoib3ZOV25MS0pjOWdZcTlvMFd1QkJrY2l5YkR4U3J3RWJyNEs3Zm81byI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770105608
6MemX0sYXu5adD6JEizcYfNgfgH0qCE3VvXvROjJ	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiVGpnaFhSUGlLNFY0OUR0SURSbTVvVEJHeTdNZnVOSnJIQVZabDJqaSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770105611
RyH5evFy8R2ftq8g6I2Kibv83OBYBGzWcKDGB8wb	\N	13.48.238.157	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiYklmTGg5c21DeVZSYlRGbzB4RDhlclhveHV5aVR1WnhaZDBpTUhGRSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770105857
Mg1BntpHImgWYdXKFfaVVdJ2TBDY2BiMf4wpfgSA	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiYUd6ZFdxN3luWG51dnpNRVAyTUFNaG1QaWxGSjlUMjNxR29ETjNjYyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770106519
amRKZl0AXJp05zCbTHp8Zbo3UCunBkK0g4rBPaIv	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiT2MxRWE0RUFqRUdTWVZYYXZMTDg5NDNaWDRLZlZEYWRjQ1ZjbmY3MiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770106520
1c8x9wdTbQ1VbqQayxzJ4AxHb4DciBQU5wM2tvXh	\N	34.155.232.229	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiTnI3QUp3d2RCWFh4ZGphM242UGZpSDJKOHk1NDE3SzY2UkpVWmluVCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770106761
bARCA5loD8IVXBuV2jKGvuzyvqZO5OB7IQJrL6vF	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZXFVMkV5VVdyaWRZREJmeXBrZXlmNGQzOFV4NFlLaFFrMDNxWVRDZyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770107431
5yADTUzhaRYgO9Bgm2QdmaFbU8DOCDxCcFzZFJzn	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiYnVIWjBWdFJ6SmVYWXhRMFlodTVwRUNGSll2MTU3M3p0Q2wzMFlFWSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770107433
bzumzQpEIBCYoQdSm7YR5cqwMo1APTg2ldxy29j9	\N	34.118.59.148	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoic2c3SVllRmpYb3h0TmVhaWV4c0sxWUhINUo3VjZxSTA4dWdUOHd3WSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770107665
a5MQ0Jtad6t2uaUIr76SfXP4H9Zy8cKvPa3HqUgY	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiTmxnR0FQSzYySjNVeEZJQm55eExVb0NUNXVLOG1Ld0hEaUNVc1p2MyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770108345
yFLCc9urrL29OQXlTxs3JZDdmOlssphjFvkf1eWJ	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiNlVZczM1NlAzbkRWbDJ0enZSb3JOd283VExNUWhOZVEzanc5SExiMCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770108348
lj8kcE3gD8RjuL902DeYnJWGK2tUq3K51KfOpho9	\N	34.65.146.108	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiWVFUSEU5SjJzUldXZldmajU4NjhUTXg0bDJoTVVjY0pOdHRWRVJNdCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770108569
L6J93wdNe0C2SW789bFH7Lf94a3xuhdI9UyT2XHx	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiUmFqdzdaSlpadUxtWFhFaEUxaGF0eFNiOFp6a0FNaDY4WFhtREtnRCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770109258
Asxs4yA8Fl9awDhry1iQXtFWf6YvUEjhazcFzMzf	\N	159.223.175.210	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiMkltT1BVMGhVMGtPYkdaeEFNb0JVWXVTcGE3aUVhVUFVNDZtNDRqeiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770076025
r2BMP4tWr347v8NCzpujg7OYlDXDnZqHj59itTnN	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiR3hwdWRIMzAxWWhIOFlpOVBLc3JKdDBJR3B5d0ZCdXdDcVZDUmRzQSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770076530
YvvedfhcwTQU0EMatyeOBhiOUWNtT537zxhIMJoa	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZDJLVXlhbTJvSDRVZzhKRVpFcGtuNnR2eVQ2VkJ6ZnJNUXlscDJVSCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770076530
ydTsX92KW1lpzhTOv6XqpyTmko30GMQOUdP4cvVV	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoid2pldDB4NEN2RkpueWpjQXdxS3F5N2JtakVDcGczNUxKYm5BT240dCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770076531
7Gq09D3AhczHUgjjzDOgeynB9A77h7NKHSnqHH5x	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoibEZHNU9MUUdSZ0o3SWN2ZFQ0UzlwT0xIUFhvZGRsSFRrQ3VqMmd1MSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770076532
komdhH435yFQ5phdfNDlh8BNbS3Hm02iadD2iTzg	\N	45.32.141.163	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoid252V2dveG1xSGd1Vk1mZWUxOUJhOThOS015dlB2WUtod3hnMG5WMSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770076930
rlSsS5TZkNsYLcf3KiAQA3LbXk4IHHMu0Tg4q0LL	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiT3FUSnFWVkxJRlJIMTJqazN0ejhRNHhJMWtCT1NNUENoNll1UGV1ayI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770077436
pDkajNjKkKsFp64kakgdMuM2O8MB2fnGaipnbRS8	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiajZrenBPSGVXWW45R2ZDSjByVjZNTzlmQlN3bU5pVjlYWnBSRXp1ZiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770077437
8CIKSBOT3Jy8f95tXc7XJe8Xcg9H9pjG3J7aUwYA	\N	143.110.177.252	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoibE5GRGJveXhEdXV4dk1TSnZiVXE4RXFCbW5VWWowQXhReWJCeEJUbiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770077836
s0mUbTE6HzX0C3htQblnT5Cdy0L8S6QeMohYi43W	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoidTVMckFLVnF6ZDIzZWtoZmVjVHdtWWpVTHM0Yk1WbFh6dWp0VW1lUiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770078343
ByTJB2bF8Zp2AhqsMtLupacprZ0KZxGecnkZCprw	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiZ0kzbms2eFZSQ0RjeU1CODB4eG9HN2hlQkxLZUphYk5hRTZLNDJuQiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770078345
jzor8XM3UW0U2rnBx0Mhy7tBD4veXVWpnpYUDcKr	\N	35.197.232.17	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiWHBLRkg1MnB0RTJCeWNCY3JESzlOTWJpTXR6NlJBZE1WWXNxSHk5dCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770078740
163qKce5cjmFPZGAsC5rClHZH15UYbUS2yaAsiNG	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiVmNTSTBoUGVGa2ZZMnk4OU5PUXplWWhLd1owODk5aEZabDlLMkdQTyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770079254
Ub0kmuHkUWUn9sY60Y9T6p4Kvl725qraFj3jPAyI	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiRFBFVUxzdmdaRHRDN3hMOHNLQ1h2Z1ZvMVR6WTNLdndOaE85QzREViI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770079256
SW98xnm6Ce2yQ0tekrY2DpExeeh3xThOZS9Ni6D0	\N	161.35.219.158	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiV3lyTzJnYzZYVGYxMDE5TDNxbnRMMHYxNU0waG9FTVNuOWNNUm00RCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770079645
ShUGnYUyWF01rQESBscVg5Xsv6idvSOn6KPh8SQd	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiM2RyYk1HaDcwMjFxaTdEQkhMa3kxVmE5aGplbzVOU1ViQlk2dVEyaCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770080180
m3nMfQD30nBQjsyyLlyGlXGZ0n89E2tqFFfAdhBk	\N	45.32.183.128	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiN0ZlU2hHYklPUVlZTkNCM1lBa0R6cjl1ajVZMXdXU1ppbXZyWlU1YiI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770080549
Am8upbMLPLQ62hZjxFn6sIgx1tULzQ2E9IaRCqSs	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoibkNteG14U2JJMVVYMGV6cnpJU0Y2d084ZGpOWnpZN0l5Y3VHTVhHTSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770081077
rbKqyz5Nsna39DgruxLskbzPs3m9NEcV99mdkFAG	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoibkFhYUpKY1dVVGxCQzh1T1p6VG5CaW1iZ1puUlBtR0paQ3FiamR1eSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770081080
qB3liW3FtJxUdxEfhbjRNFF7uRwKeDyCB1278b6d	\N	34.94.216.215	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiQTZFNzk4Nk0wNWdTNkIwZU9FQmE5TUpNZWt5b3puQWxjWm1WNlNxVCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770081454
uU7U53K2GqBwLRWs9gZJtkToy23IRwbmR0F1K4iP	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoic25mUzJLQ0pTdmhXWWRIbGRrZVdZSEFwSUVJWWhSd2U2MmdxeHFsVyI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770081987
3yHQEdlLEfe3D93XCa9Qk5byPhHI0HVI4I3DzWTk	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoicVN3b0NSQUtKZ0VLa00yNFpvV3Vla1ZKd0djMkRiSVFJQ2lWeHVMYSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770081990
apc18Jb0jJfJ1rY2fUHgIjPPce2nxZ6rzsC7P1ti	\N	45.63.61.213	Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/129.0.0.0 Safari/537.36 (StatusCake)	YTozOntzOjY6Il90b2tlbiI7czo0MDoiYjlrdzZqRTVjaVc4YXlvVmdDNUpXZHJwUzdUbDY5OXNpUWpaQ092eSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770082356
e6D9oHqyKmuVzUJkcLinGNwjbleTV60HcMLKx5dN	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiN3k2QmNKd213TGhVUFRyNnhkU1VDeldEZmZnSzdUZEJzalFVaDlhaCI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770082902
8UH9ZGHJP54EBmJoefF6xhuLfKXZdJx3zF6tXA3e	\N	207.154.246.185	StatusCake_Pagespeed_Indev	YTozOntzOjY6Il90b2tlbiI7czo0MDoiU2txM0VBb1NmV0J0OHZPaHY2YTAzSExRcEtreG4yWmFhMU1rYTZUQSI7czo5OiJfcHJldmlvdXMiO2E6Mjp7czozOiJ1cmwiO3M6MzM6Imh0dHBzOi8va2VnZWxtYXN0ZXIuc2NobnVyYnVzLm5ldCI7czo1OiJyb3V0ZSI7czo0OiJob21lIjt9czo2OiJfZmxhc2giO2E6Mjp7czozOiJvbGQiO2E6MDp7fXM6MzoibmV3IjthOjA6e319fQ==	1770082903
\.


--
-- Data for Name: transactions; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.transactions (id, club_id, player_id, matchday_id, fee_entry_id, type, amount, date, notes, created_at, updated_at) FROM stdin;
201	1	1	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
202	1	10	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
203	1	8	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
204	1	4	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
205	1	9	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
206	1	3	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
207	1	7	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
208	1	2	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
209	1	11	17	\N	1	-1000	2025-01-24	\N	2025-03-17 14:12:31	2025-03-17 14:12:31
281	1	10	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
282	1	3	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
283	1	1	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
284	1	7	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
285	1	8	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
286	1	9	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
287	1	4	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
288	1	11	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
289	1	2	18	\N	1	-1000	2024-09-06	\N	2025-03-18 07:05:25	2025-03-18 07:05:25
290	1	3	18	615	2	-20	2024-09-06	\N	2025-03-18 07:05:47	2025-03-18 07:05:47
291	1	3	18	613	2	-40	2024-09-06	\N	2025-03-18 07:05:47	2025-03-18 07:05:47
292	1	8	18	617	2	-140	2024-09-06	\N	2025-03-18 07:06:12	2025-03-18 07:06:12
293	1	8	18	618	2	-120	2024-09-06	\N	2025-03-18 07:06:12	2025-03-18 07:06:12
294	1	8	18	619	2	-40	2024-09-06	\N	2025-03-18 07:06:13	2025-03-18 07:06:13
295	1	4	18	625	2	-120	2024-09-06	\N	2025-03-18 07:07:16	2025-03-18 07:07:16
296	1	4	18	626	2	-60	2024-09-06	\N	2025-03-18 07:07:16	2025-03-18 07:07:16
297	1	10	18	665	2	-100	2024-09-06	\N	2025-03-18 07:07:23	2025-03-18 07:07:23
298	1	10	18	666	2	-120	2024-09-06	\N	2025-03-18 07:07:23	2025-03-18 07:07:23
299	1	9	18	633	2	-120	2024-09-06	\N	2025-03-18 07:07:31	2025-03-18 07:07:31
300	1	9	18	634	2	-80	2024-09-06	\N	2025-03-18 07:07:31	2025-03-18 07:07:31
301	1	2	18	649	2	-60	2024-09-06	\N	2025-03-18 07:07:41	2025-03-18 07:07:41
302	1	2	18	650	2	-100	2024-09-06	\N	2025-03-18 07:07:41	2025-03-18 07:07:41
303	1	1	18	657	2	-20	2024-09-06	\N	2025-03-18 07:07:57	2025-03-18 07:07:57
304	1	1	18	658	2	-140	2024-09-06	\N	2025-03-18 07:07:58	2025-03-18 07:07:58
305	1	11	18	641	2	-60	2024-09-06	\N	2025-03-18 07:08:08	2025-03-18 07:08:08
306	1	11	18	642	2	-180	2024-09-06	\N	2025-03-18 07:08:09	2025-03-18 07:08:09
307	1	1	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
308	1	7	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
309	1	3	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
310	1	8	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
311	1	4	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
312	1	10	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
313	1	9	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
516	1	\N	\N	\N	5	-3000	2025-02-21	Bahn	2025-03-20 12:46:54	2025-03-20 12:46:54
210	1	10	17	535	2	-80	2025-01-24	\N	2025-03-17 14:14:51	2025-03-17 14:14:51
211	1	10	17	533	2	-100	2025-01-24	\N	2025-03-17 14:14:51	2025-03-17 14:14:51
212	1	8	17	543	2	-140	2025-01-24	\N	2025-03-17 14:15:22	2025-03-17 14:15:22
213	1	8	17	541	2	-80	2025-01-24	\N	2025-03-17 14:15:23	2025-03-17 14:15:23
718	1	7	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:44	2025-06-16 12:10:44
719	1	8	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:44	2025-06-16 12:10:44
720	1	10	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:44	2025-06-16 12:10:44
721	1	9	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:44	2025-06-16 12:10:44
722	1	2	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:44	2025-06-16 12:10:44
723	1	11	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:44	2025-06-16 12:10:44
724	1	3	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:44	2025-06-16 12:10:44
725	1	4	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:45	2025-06-16 12:10:45
726	1	1	55	\N	1	-1000	2025-06-13	\N	2025-06-16 12:10:45	2025-06-16 12:10:45
727	1	10	55	2181	2	-40	2025-06-13	\N	2025-06-16 12:12:32	2025-06-16 12:12:32
728	1	10	55	2182	2	-40	2025-06-13	\N	2025-06-16 12:12:32	2025-06-16 12:12:32
729	1	8	55	2189	2	-80	2025-06-13	\N	2025-06-16 12:13:05	2025-06-16 12:13:05
730	1	8	55	2190	2	-20	2025-06-13	\N	2025-06-16 12:13:07	2025-06-16 12:13:07
731	1	4	55	2197	2	-160	2025-06-13	\N	2025-06-16 12:13:24	2025-06-16 12:13:24
732	1	4	55	2198	2	-60	2025-06-13	\N	2025-06-16 12:13:24	2025-06-16 12:13:24
806	1	1	\N	\N	4	260	2025-07-11	\N	2025-08-08 15:09:48	2025-08-08 15:09:48
807	1	\N	\N	\N	5	-9000	2025-08-08	Kegelbahngebühr	2025-09-05 15:06:54	2025-09-05 15:07:26
808	1	2	\N	\N	3	1160	2025-08-08	\N	2025-09-05 15:08:27	2025-09-05 15:08:27
809	1	1	\N	\N	3	1260	2025-08-08	\N	2025-09-05 15:08:27	2025-09-05 15:08:27
179	1	3	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
810	1	2	\N	\N	4	80	2025-08-08	Auto Tip	2025-09-05 15:08:27	2025-09-05 15:08:27
811	1	7	\N	\N	3	2280	2025-08-08	\N	2025-09-05 15:09:31	2025-09-05 15:09:31
812	1	7	\N	\N	4	220	2025-08-08	\N	2025-09-05 15:09:31	2025-09-05 15:09:31
813	1	9	\N	\N	3	2300	2025-08-08	\N	2025-09-05 15:10:02	2025-09-05 15:10:02
814	1	8	\N	\N	3	2100	2025-08-08	\N	2025-09-05 15:10:02	2025-09-05 15:10:02
815	1	9	\N	\N	4	100	2025-08-08	Auto Tip	2025-09-05 15:10:02	2025-09-05 15:10:02
816	1	11	\N	\N	3	2200	2025-08-08	\N	2025-09-05 15:11:23	2025-09-05 15:11:23
817	1	10	\N	\N	3	2080	2025-08-08	\N	2025-09-05 15:11:23	2025-09-05 15:11:23
818	1	11	\N	\N	4	120	2025-08-08	Auto Tip	2025-09-05 15:11:23	2025-09-05 15:11:23
819	1	3	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:36	2025-09-05 15:12:36
820	1	2	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:37	2025-09-05 15:12:37
821	1	7	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:37	2025-09-05 15:12:37
180	1	1	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
181	1	7	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
182	1	4	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
183	1	8	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
184	1	10	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
185	1	9	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
186	1	2	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
187	1	11	16	\N	1	-1000	2024-11-29	\N	2025-03-17 13:59:42	2025-03-17 13:59:42
188	1	7	16	487	2	-140	2024-11-29	\N	2025-03-17 14:01:46	2025-03-17 14:01:46
189	1	7	16	485	2	-160	2024-11-29	\N	2025-03-17 14:01:46	2025-03-17 14:01:46
190	1	8	16	495	2	-60	2024-11-29	\N	2025-03-17 14:02:24	2025-03-17 14:02:24
191	1	8	16	493	2	-120	2024-11-29	\N	2025-03-17 14:02:24	2025-03-17 14:02:24
192	1	4	16	503	2	-180	2024-11-29	\N	2025-03-17 14:02:57	2025-03-17 14:02:57
193	1	4	16	501	2	-120	2024-11-29	\N	2025-03-17 14:02:57	2025-03-17 14:02:57
194	1	9	16	511	2	-180	2024-11-29	\N	2025-03-17 14:03:20	2025-03-17 14:03:20
195	1	9	16	509	2	-260	2024-11-29	\N	2025-03-17 14:03:22	2025-03-17 14:03:22
196	1	2	16	519	2	-80	2024-11-29	\N	2025-03-17 14:03:40	2025-03-17 14:03:40
197	1	2	16	517	2	-140	2024-11-29	\N	2025-03-17 14:03:41	2025-03-17 14:03:41
198	1	3	16	527	2	-80	2024-11-29	\N	2025-03-17 14:04:03	2025-03-17 14:04:03
199	1	3	16	525	2	-80	2024-11-29	\N	2025-03-17 14:04:04	2025-03-17 14:04:04
200	1	9	16	506	2	-100	2024-11-29	\N	2025-03-17 14:05:40	2025-03-17 14:05:40
214	1	8	17	537	2	-20	2025-01-24	\N	2025-03-17 14:15:23	2025-03-17 14:15:23
215	1	12	17	551	2	-40	2025-01-24	\N	2025-03-17 14:15:42	2025-03-17 14:15:42
216	1	12	17	549	2	-120	2025-01-24	\N	2025-03-17 14:15:43	2025-03-17 14:15:43
217	1	4	17	559	2	-220	2025-01-24	\N	2025-03-17 14:16:01	2025-03-17 14:16:01
218	1	4	17	557	2	-100	2025-01-24	\N	2025-03-17 14:16:02	2025-03-17 14:16:02
219	1	9	17	567	2	-120	2025-01-24	\N	2025-03-17 14:16:26	2025-03-17 14:16:26
220	1	9	17	565	2	-60	2025-01-24	\N	2025-03-17 14:16:27	2025-03-17 14:16:27
221	1	4	17	556	2	-100	2025-01-24	\N	2025-03-17 14:16:37	2025-03-17 14:16:37
222	1	13	17	575	2	-140	2025-01-24	\N	2025-03-17 14:17:04	2025-03-17 14:17:04
223	1	13	17	573	2	-100	2025-01-24	\N	2025-03-17 14:17:05	2025-03-17 14:17:05
224	1	2	17	583	2	-140	2025-01-24	\N	2025-03-17 14:17:21	2025-03-17 14:17:21
225	1	2	17	581	2	-140	2025-01-24	\N	2025-03-17 14:17:22	2025-03-17 14:17:22
226	1	7	17	591	2	-200	2025-01-24	\N	2025-03-17 14:17:41	2025-03-17 14:17:41
227	1	7	17	589	2	-180	2025-01-24	\N	2025-03-17 14:17:41	2025-03-17 14:17:41
228	1	1	17	599	2	-80	2025-01-24	\N	2025-03-17 14:18:00	2025-03-17 14:18:00
229	1	1	17	597	2	-160	2025-01-24	\N	2025-03-17 14:18:01	2025-03-17 14:18:01
230	1	3	17	607	2	-60	2025-01-24	\N	2025-03-17 14:18:15	2025-03-17 14:18:15
231	1	3	17	605	2	-60	2025-01-24	\N	2025-03-17 14:18:16	2025-03-17 14:18:16
314	1	11	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
315	1	2	19	\N	1	-1000	2024-07-12	\N	2025-03-18 07:09:20	2025-03-18 07:09:20
316	1	7	19	687	2	-120	2024-07-12	\N	2025-03-18 07:10:07	2025-03-18 07:10:07
317	1	7	19	685	2	-120	2024-07-12	\N	2025-03-18 07:10:07	2025-03-18 07:10:07
318	1	7	19	682	2	-200	2024-07-12	\N	2025-03-18 07:10:08	2025-03-18 07:10:08
319	1	10	19	679	2	-120	2024-07-12	\N	2025-03-18 07:10:15	2025-03-18 07:10:15
320	1	10	19	677	2	-140	2024-07-12	\N	2025-03-18 07:10:16	2025-03-18 07:10:16
321	1	2	19	703	2	-160	2024-07-12	\N	2025-03-18 07:10:31	2025-03-18 07:10:31
322	1	2	19	701	2	-80	2024-07-12	\N	2025-03-18 07:10:31	2025-03-18 07:10:31
323	1	2	19	698	2	-200	2024-07-12	\N	2025-03-18 07:10:31	2025-03-18 07:10:31
324	1	1	19	711	2	-160	2024-07-12	\N	2025-03-18 07:10:41	2025-03-18 07:10:41
325	1	1	19	709	2	-140	2024-07-12	\N	2025-03-18 07:10:42	2025-03-18 07:10:42
326	1	1	19	705	2	-20	2024-07-12	\N	2025-03-18 07:10:43	2025-03-18 07:10:43
327	1	11	19	695	2	-280	2024-07-12	\N	2025-03-18 07:10:51	2025-03-18 07:10:51
328	1	11	19	693	2	-300	2024-07-12	\N	2025-03-18 07:10:52	2025-03-18 07:10:52
329	1	11	19	689	2	-40	2024-07-12	\N	2025-03-18 07:10:53	2025-03-18 07:10:53
330	1	3	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
331	1	8	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
332	1	4	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
333	1	9	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
334	1	7	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
335	1	10	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
336	1	1	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
337	1	2	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
338	1	11	20	\N	1	-1000	2024-06-16	\N	2025-03-18 07:11:54	2025-03-18 07:11:54
339	1	4	20	735	2	-160	2024-06-16	\N	2025-03-18 07:12:47	2025-03-18 07:12:47
340	1	4	20	733	2	-160	2024-06-16	\N	2025-03-18 07:12:48	2025-03-18 07:12:48
341	1	4	20	729	2	-40	2024-06-16	\N	2025-03-18 07:12:49	2025-03-18 07:12:49
342	1	8	20	727	2	-60	2024-06-16	\N	2025-03-18 07:12:58	2025-03-18 07:13:04
343	1	8	20	725	2	-40	2024-06-16	\N	2025-03-18 07:13:05	2025-03-18 07:13:05
344	1	8	20	721	2	-20	2024-06-16	\N	2025-03-18 07:13:06	2025-03-18 07:13:06
345	1	10	20	719	2	-80	2024-06-16	\N	2025-03-18 07:13:15	2025-03-18 07:13:15
346	1	10	20	717	2	-100	2024-06-16	\N	2025-03-18 07:13:15	2025-03-18 07:13:15
347	1	2	20	751	2	-140	2024-06-16	\N	2025-03-18 07:13:35	2025-03-18 07:13:35
348	1	2	20	749	2	-120	2024-06-16	\N	2025-03-18 07:13:36	2025-03-18 07:13:36
349	1	11	20	743	2	-140	2024-06-16	\N	2025-03-18 07:13:46	2025-03-18 07:13:46
350	1	11	20	741	2	-180	2024-06-16	\N	2025-03-18 07:13:46	2025-03-18 07:13:46
351	1	11	20	737	2	-20	2024-06-16	\N	2025-03-18 07:13:47	2025-03-18 07:13:47
352	1	3	20	759	2	-40	2024-06-16	\N	2025-03-18 07:13:58	2025-03-18 07:13:58
353	1	3	20	757	2	-40	2024-06-16	\N	2025-03-18 07:13:58	2025-03-18 07:13:58
354	1	9	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
355	1	7	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
356	1	1	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
357	1	4	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
358	1	8	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
359	1	10	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
360	1	3	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
361	1	2	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
362	1	11	21	\N	1	-1000	2024-04-19	\N	2025-03-18 07:15:42	2025-03-18 07:15:42
363	1	7	21	799	2	-100	2024-04-19	\N	2025-03-18 07:16:22	2025-03-18 07:16:22
364	1	7	21	797	2	-120	2024-04-19	\N	2025-03-18 07:16:22	2025-03-18 07:16:22
365	1	10	21	767	2	-60	2024-04-19	\N	2025-03-18 07:16:39	2025-03-18 07:16:39
366	1	10	21	766	2	-500	2024-04-19	\N	2025-03-18 07:16:39	2025-03-18 07:16:39
367	1	2	21	783	2	-100	2024-04-19	\N	2025-03-18 07:16:49	2025-03-18 07:16:49
368	1	2	21	781	2	-20	2024-04-19	\N	2025-03-18 07:16:49	2025-03-18 07:16:49
369	1	1	21	791	2	-180	2024-04-19	\N	2025-03-18 07:17:00	2025-03-18 07:17:00
370	1	1	21	789	2	-160	2024-04-19	\N	2025-03-18 07:17:01	2025-03-18 07:17:01
371	1	1	21	790	2	-500	2024-04-19	\N	2025-03-18 07:17:01	2025-03-18 07:17:01
372	1	11	21	775	2	-160	2024-04-19	\N	2025-03-18 07:17:13	2025-03-18 07:17:13
373	1	11	21	773	2	-240	2024-04-19	\N	2025-03-18 07:17:14	2025-03-18 07:17:14
374	1	11	21	769	2	-20	2024-04-19	\N	2025-03-18 07:17:15	2025-03-18 07:17:15
375	1	11	21	774	2	-500	2024-04-19	\N	2025-03-18 07:17:16	2025-03-18 07:17:16
376	1	7	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
377	1	10	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
378	1	1	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
379	1	9	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
380	1	4	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
381	1	8	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
382	1	3	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
383	1	2	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
384	1	11	22	\N	1	-1000	2024-03-22	\N	2025-03-18 07:18:17	2025-03-18 07:18:17
385	1	14	22	823	2	-140	2024-03-22	\N	2025-03-18 07:19:40	2025-03-18 07:19:40
386	1	14	22	821	2	-60	2024-03-22	\N	2025-03-18 07:19:41	2025-03-18 07:19:41
387	1	10	22	807	2	-40	2024-03-22	\N	2025-03-18 07:20:17	2025-03-18 07:20:17
388	1	10	22	805	2	-100	2024-03-22	\N	2025-03-18 07:20:17	2025-03-18 07:20:17
389	1	9	22	815	2	-140	2024-03-22	\N	2025-03-18 07:20:29	2025-03-18 07:20:29
390	1	9	22	813	2	-120	2024-03-22	\N	2025-03-18 07:20:29	2025-03-18 07:20:29
391	1	9	22	814	2	-500	2024-03-22	\N	2025-03-18 07:20:29	2025-03-18 07:20:29
392	1	2	22	831	2	-120	2024-03-22	\N	2025-03-18 07:20:42	2025-03-18 07:20:42
393	1	2	22	829	2	-100	2024-03-22	\N	2025-03-18 07:20:43	2025-03-18 07:20:43
394	1	1	22	839	2	-100	2024-03-22	\N	2025-03-18 07:20:51	2025-03-18 07:20:51
395	1	1	22	837	2	-160	2024-03-22	\N	2025-03-18 07:20:51	2025-03-18 07:20:51
396	1	1	22	833	2	-20	2024-03-22	\N	2025-03-18 07:20:51	2025-03-18 07:20:51
415	1	4	23	855	2	-180	2024-10-04	\N	2025-03-18 20:20:16	2025-03-18 20:20:16
406	1	8	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
407	1	4	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
408	1	9	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
409	1	7	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
410	1	3	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
411	1	1	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
412	1	10	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
413	1	11	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
414	1	2	23	\N	1	-1000	2024-10-04	\N	2025-03-18 20:18:39	2025-03-18 20:18:39
416	1	4	23	853	2	-200	2024-10-04	\N	2025-03-18 20:20:16	2025-03-18 20:20:16
417	1	7	23	911	2	-160	2024-10-04	\N	2025-03-18 20:20:27	2025-03-18 20:20:27
418	1	7	23	909	2	-100	2024-10-04	\N	2025-03-18 20:20:28	2025-03-18 20:20:28
419	1	8	23	863	2	-40	2024-10-04	\N	2025-03-18 20:20:35	2025-03-18 20:20:35
420	1	8	23	861	2	-80	2024-10-04	\N	2025-03-18 20:20:36	2025-03-18 20:20:36
421	1	12	23	919	2	-80	2024-10-04	\N	2025-03-18 20:20:44	2025-03-18 20:20:44
422	1	12	23	917	2	-100	2024-10-04	\N	2025-03-18 20:20:45	2025-03-18 20:20:45
423	1	10	23	887	2	-40	2024-10-04	\N	2025-03-18 20:20:55	2025-03-18 20:20:55
424	1	10	23	885	2	-80	2024-10-04	\N	2025-03-18 20:20:56	2025-03-18 20:20:56
425	1	9	23	903	2	-160	2024-10-04	\N	2025-03-18 20:21:12	2025-03-18 20:21:12
426	1	9	23	901	2	-60	2024-10-04	\N	2025-03-18 20:21:13	2025-03-18 20:21:13
427	1	9	23	899	2	-20	2024-10-04	\N	2025-03-18 20:21:14	2025-03-18 20:21:14
428	1	9	23	902	2	-500	2024-10-04	\N	2025-03-18 20:21:15	2025-03-18 20:21:15
429	1	11	23	879	2	-120	2024-10-04	\N	2025-03-18 20:21:30	2025-03-18 20:21:30
430	1	11	23	877	2	-300	2024-10-04	\N	2025-03-18 20:21:31	2025-03-18 20:21:31
431	1	11	23	873	2	-60	2024-10-04	\N	2025-03-18 20:21:32	2025-03-18 20:21:32
432	1	13	23	895	2	-160	2024-10-04	\N	2025-03-18 20:21:43	2025-03-18 20:21:43
433	1	13	23	893	2	-100	2024-10-04	\N	2025-03-18 20:21:44	2025-03-18 20:21:44
434	1	3	23	871	2	-60	2024-10-04	\N	2025-03-18 20:21:52	2025-03-18 20:21:52
435	1	3	23	869	2	-60	2024-10-04	\N	2025-03-18 20:21:52	2025-03-18 20:21:52
436	1	8	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
437	1	4	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
438	1	9	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
439	1	1	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
440	1	3	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
441	1	10	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
442	1	7	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
443	1	11	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
444	1	2	24	\N	1	-1000	2024-12-27	\N	2025-03-18 20:24:11	2025-03-18 20:24:11
445	1	4	24	951	2	-140	2024-12-27	\N	2025-03-18 20:25:14	2025-03-18 20:25:14
446	1	4	24	949	2	-100	2024-12-27	\N	2025-03-18 20:25:14	2025-03-18 20:25:14
447	1	7	24	959	2	-60	2024-12-27	\N	2025-03-18 20:25:23	2025-03-18 20:25:23
448	1	7	24	957	2	-80	2024-12-27	\N	2025-03-18 20:25:24	2025-03-18 20:25:24
449	1	7	24	958	2	-500	2024-12-27	\N	2025-03-18 20:25:25	2025-03-18 20:25:25
450	1	8	24	967	2	-100	2024-12-27	\N	2025-03-18 20:25:36	2025-03-18 20:25:36
451	1	8	24	965	2	-100	2024-12-27	\N	2025-03-18 20:25:36	2025-03-18 20:25:36
452	1	8	24	961	2	-20	2024-12-27	\N	2025-03-18 20:25:36	2025-03-18 20:25:36
453	1	10	24	975	2	-120	2024-12-27	\N	2025-03-18 20:25:46	2025-03-18 20:25:46
454	1	10	24	973	2	-180	2024-12-27	\N	2025-03-18 20:25:46	2025-03-18 20:25:46
455	1	9	24	977	2	-80	2024-12-27	\N	2025-03-18 20:25:55	2025-03-18 20:25:55
456	1	9	24	978	2	-120	2024-12-27	\N	2025-03-18 20:25:56	2025-03-18 20:25:56
457	1	2	24	943	2	-60	2024-12-27	\N	2025-03-18 20:26:06	2025-03-18 20:26:06
458	1	2	24	941	2	-100	2024-12-27	\N	2025-03-18 20:26:06	2025-03-18 20:26:06
459	1	1	24	935	2	-100	2024-12-27	\N	2025-03-18 20:26:16	2025-03-18 20:26:16
460	1	1	24	933	2	-160	2024-12-27	\N	2025-03-18 20:26:17	2025-03-18 20:26:17
461	1	3	24	925	2	-20	2024-12-27	\N	2025-03-18 20:26:28	2025-03-18 20:26:28
517	1	4	\N	\N	3	10880	2025-02-21	Initialer Ausgleich	2025-03-21 16:32:50	2025-03-21 16:32:50
518	1	8	\N	\N	3	10180	2025-02-21	Initialer Ausgleich	2025-03-21 16:33:19	2025-03-21 16:33:19
519	1	12	\N	\N	3	340	2025-02-21	Initialer Ausgleich	2025-03-21 16:33:47	2025-03-21 16:33:47
520	1	10	\N	\N	3	10960	2025-02-21	Initialer Ausgleich	2025-03-21 16:34:06	2025-03-21 16:34:06
521	1	9	\N	\N	3	11620	2025-02-21	Initialer Ausgleich	2025-03-21 16:34:25	2025-03-21 16:34:25
522	1	2	\N	\N	3	10860	2025-02-21	Initialer Ausgleich	2025-03-21 16:34:45	2025-03-21 16:34:45
523	1	1	\N	\N	3	11100	2025-02-21	Initialer Ausgleich	2025-03-21 16:35:04	2025-03-21 16:35:04
524	1	11	\N	\N	3	11600	2025-02-21	Initialer Ausgleich	2025-03-21 16:35:23	2025-03-21 16:35:23
525	1	13	\N	\N	3	500	2025-02-21	Initialer Ausgleich	2025-03-21 16:35:50	2025-03-21 16:35:50
526	1	3	\N	\N	3	9560	2025-02-21	Initialer Ausgleich	2025-03-21 16:36:08	2025-03-21 16:36:08
527	1	7	\N	\N	3	7860	2025-01-24	Initialer Ausgleich	2025-03-21 16:38:04	2025-03-21 16:38:04
528	1	4	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:41	2025-03-21 16:38:41
529	1	8	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:41	2025-03-21 16:38:41
530	1	10	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:42	2025-03-21 16:38:42
531	1	9	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:43	2025-03-21 16:38:43
532	1	2	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:43	2025-03-21 16:38:43
533	1	1	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:45	2025-03-21 16:38:45
534	1	11	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:45	2025-03-21 16:38:45
535	1	3	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:45	2025-03-21 16:38:45
536	1	7	31	\N	1	-1000	2025-02-21	\N	2025-03-21 16:38:46	2025-03-21 16:38:46
537	1	10	31	1300	2	-40	2025-02-21	\N	2025-03-21 16:40:01	2025-03-21 16:40:01
538	1	10	31	1301	2	-60	2025-02-21	\N	2025-03-21 16:40:01	2025-03-21 16:40:01
539	1	8	31	1236	2	-140	2025-02-21	\N	2025-03-21 16:40:15	2025-03-21 16:40:15
540	1	8	31	1237	2	-220	2025-02-21	\N	2025-03-21 16:40:17	2025-03-21 16:40:17
541	1	12	31	1244	2	-20	2025-02-21	\N	2025-03-21 16:40:24	2025-03-21 16:40:24
542	1	12	31	1245	2	-80	2025-02-21	\N	2025-03-21 16:40:24	2025-03-21 16:40:24
543	1	4	31	1228	2	-180	2025-02-21	\N	2025-03-21 16:40:39	2025-03-21 16:40:39
544	1	4	31	1229	2	-140	2025-02-21	\N	2025-03-21 16:40:40	2025-03-21 16:40:40
545	1	4	31	1230	2	-20	2025-02-21	\N	2025-03-21 16:40:41	2025-03-21 16:40:41
546	1	9	31	1252	2	-140	2025-02-21	\N	2025-03-21 16:40:52	2025-03-21 16:40:52
547	1	9	31	1253	2	-80	2025-02-21	\N	2025-03-21 16:40:53	2025-03-21 16:40:53
548	1	13	31	1284	2	-120	2025-02-21	\N	2025-03-21 16:41:08	2025-03-21 16:41:08
549	1	13	31	1285	2	-60	2025-02-21	\N	2025-03-21 16:41:08	2025-03-21 16:41:08
550	1	11	31	1276	2	-120	2025-02-21	\N	2025-03-21 16:41:29	2025-03-21 16:41:29
551	1	11	31	1277	2	-240	2025-02-21	\N	2025-03-21 16:41:30	2025-03-21 16:41:30
552	1	2	31	1260	2	-120	2025-02-21	\N	2025-03-21 16:41:55	2025-03-21 16:41:55
553	1	2	31	1261	2	-120	2025-02-21	\N	2025-03-21 16:41:55	2025-03-21 16:41:55
554	1	1	31	1268	2	-100	2025-02-21	\N	2025-03-21 16:42:04	2025-03-21 16:42:04
555	1	1	31	1269	2	-40	2025-02-21	\N	2025-03-21 16:42:05	2025-03-21 16:42:05
556	1	1	31	1270	2	-20	2025-02-21	\N	2025-03-21 16:42:05	2025-03-21 16:42:05
557	1	3	31	1293	2	-20	2025-02-21	\N	2025-03-21 16:42:11	2025-03-21 16:42:11
733	1	9	55	2205	2	-100	2025-06-13	\N	2025-06-16 12:13:55	2025-06-16 12:13:55
734	1	9	55	2206	2	-100	2025-06-13	\N	2025-06-16 12:13:56	2025-06-16 12:13:56
735	1	9	55	2209	2	-100	2025-06-13	\N	2025-06-16 12:13:57	2025-06-16 12:13:57
736	1	11	55	2213	2	-80	2025-06-13	\N	2025-06-16 12:14:31	2025-06-16 12:14:31
737	1	11	55	2214	2	-120	2025-06-13	\N	2025-06-16 12:14:33	2025-06-16 12:14:33
738	1	2	55	2221	2	-40	2025-06-13	\N	2025-06-16 12:14:48	2025-06-16 12:14:48
739	1	2	55	2222	2	-20	2025-06-13	\N	2025-06-16 12:14:49	2025-06-16 12:14:49
740	1	1	55	2229	2	-100	2025-06-13	\N	2025-06-16 12:15:06	2025-06-16 12:15:06
741	1	1	55	2230	2	-80	2025-06-13	\N	2025-06-16 12:15:07	2025-06-16 12:15:07
742	1	3	55	2237	2	-40	2025-06-13	\N	2025-06-16 12:15:33	2025-06-16 12:15:33
743	1	3	55	2238	2	-60	2025-06-13	\N	2025-06-16 12:15:33	2025-06-16 12:15:33
744	1	3	55	2244	2	-1000	2025-06-13	\N	2025-06-16 12:15:34	2025-06-16 12:15:34
745	1	2	\N	\N	3	2340	2025-06-13	\N	2025-06-16 12:16:53	2025-06-16 12:16:53
746	1	1	\N	\N	3	2360	2025-06-13	\N	2025-06-16 12:16:53	2025-06-16 12:16:53
747	1	2	\N	\N	4	300	2025-06-13	Auto Tip	2025-06-16 12:16:53	2025-06-16 12:16:53
748	1	3	\N	\N	3	3100	2025-06-13	\N	2025-06-16 12:17:34	2025-06-16 12:17:34
749	1	4	\N	\N	3	2440	2025-06-13	\N	2025-06-16 12:17:34	2025-06-16 12:17:34
750	1	3	\N	\N	4	460	2025-06-13	Auto Tip	2025-06-16 12:17:34	2025-06-16 12:17:34
751	1	8	\N	\N	3	2800	2025-06-13	\N	2025-06-16 12:18:07	2025-06-16 12:18:07
752	1	9	\N	\N	3	2320	2025-06-13	\N	2025-06-16 12:18:07	2025-06-16 12:18:07
753	1	8	\N	\N	4	380	2025-06-13	Auto Tip	2025-06-16 12:18:07	2025-06-16 12:18:07
754	1	10	\N	\N	3	2200	2025-06-13	\N	2025-06-16 12:18:41	2025-06-16 12:18:41
755	1	11	\N	\N	3	2340	2025-06-13	\N	2025-06-16 12:18:41	2025-06-16 12:18:41
756	1	10	\N	\N	4	460	2025-06-13	Auto Tip	2025-06-16 12:18:41	2025-06-16 12:18:41
588	1	7	\N	\N	3	4380	2025-03-21	\N	2025-06-13 15:14:30	2025-06-13 15:14:30
589	1	7	\N	\N	4	120	2025-03-21	\N	2025-06-13 15:14:30	2025-06-13 15:14:30
590	1	9	\N	\N	3	1220	2025-03-21	\N	2025-06-13 15:15:54	2025-06-13 15:15:54
591	1	8	\N	\N	3	1360	2025-03-21	\N	2025-06-13 15:15:54	2025-06-13 15:15:54
592	1	9	\N	\N	4	120	2025-03-21	Auto Tip	2025-06-13 15:15:54	2025-06-13 15:15:54
593	1	13	\N	\N	3	180	2025-03-21	\N	2025-06-13 15:16:56	2025-06-13 15:16:56
594	1	12	\N	\N	3	100	2025-03-21	\N	2025-06-13 15:16:56	2025-06-13 15:16:56
595	1	13	\N	\N	4	220	2025-03-21	Auto Tip	2025-06-13 15:16:56	2025-06-13 15:16:56
596	1	10	\N	\N	3	1100	2025-03-21	\N	2025-06-13 15:17:32	2025-06-13 15:17:32
597	1	11	\N	\N	3	1360	2025-03-21	\N	2025-06-13 15:17:32	2025-06-13 15:17:32
598	1	10	\N	\N	4	40	2025-03-21	Auto Tip	2025-06-13 15:17:32	2025-06-13 15:17:32
599	1	2	\N	\N	3	1240	2025-03-21	\N	2025-06-13 15:18:05	2025-06-13 15:18:05
600	1	1	\N	\N	3	1160	2025-03-21	\N	2025-06-13 15:18:05	2025-06-13 15:18:05
601	1	2	\N	\N	4	100	2025-03-21	Auto Tip	2025-06-13 15:18:05	2025-06-13 15:18:05
602	1	3	\N	\N	3	1020	2025-03-21	\N	2025-06-13 15:18:39	2025-06-13 15:18:39
603	1	4	\N	\N	3	1340	2025-03-21	\N	2025-06-13 15:18:39	2025-06-13 15:18:39
604	1	3	\N	\N	4	140	2025-03-21	Auto Tip	2025-06-13 15:18:39	2025-06-13 15:18:39
649	1	11	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:01	2025-06-13 15:34:01
650	1	10	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:01	2025-06-13 15:34:01
651	1	8	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:02	2025-06-13 15:34:02
652	1	9	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:02	2025-06-13 15:34:02
653	1	2	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:03	2025-06-13 15:34:03
654	1	7	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:04	2025-06-13 15:34:04
655	1	3	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:04	2025-06-13 15:34:04
656	1	1	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:05	2025-06-13 15:34:05
658	1	4	53	\N	1	-1000	2025-03-21	\N	2025-06-13 15:34:05	2025-06-13 15:34:05
659	1	10	53	2045	2	-60	2025-03-21	\N	2025-06-13 15:36:42	2025-06-13 15:36:42
660	1	10	53	2046	2	-40	2025-03-21	\N	2025-06-13 15:36:42	2025-06-13 15:36:42
661	1	8	53	2053	2	-80	2025-03-21	\N	2025-06-13 15:36:51	2025-06-13 15:36:51
662	1	8	53	2054	2	-40	2025-03-21	\N	2025-06-13 15:36:51	2025-06-13 15:36:51
663	1	8	53	2055	2	-40	2025-03-21	\N	2025-06-13 15:36:51	2025-06-13 15:36:51
664	1	8	53	2060	2	-500	2025-03-21	\N	2025-06-13 15:36:51	2025-06-13 15:36:51
665	1	4	53	2061	2	-140	2025-03-21	\N	2025-06-13 15:37:46	2025-06-13 15:37:46
666	1	4	53	2062	2	-60	2025-03-21	\N	2025-06-13 15:37:47	2025-06-13 15:37:47
667	1	4	53	2063	2	-40	2025-03-21	\N	2025-06-13 15:37:48	2025-06-13 15:37:48
668	1	9	53	2069	2	-100	2025-03-21	\N	2025-06-13 15:38:01	2025-06-13 15:38:01
669	1	9	53	2070	2	-40	2025-03-21	\N	2025-06-13 15:38:02	2025-06-13 15:38:02
670	1	2	53	2077	2	-100	2025-03-21	\N	2025-06-13 15:38:11	2025-06-13 15:38:11
671	1	2	53	2078	2	-60	2025-03-21	\N	2025-06-13 15:38:11	2025-06-13 15:38:11
672	1	7	53	2085	2	-40	2025-03-21	\N	2025-06-13 15:38:18	2025-06-13 15:38:18
673	1	7	53	2086	2	-80	2025-03-21	\N	2025-06-13 15:38:19	2025-06-13 15:38:19
674	1	1	53	2093	2	-80	2025-03-21	\N	2025-06-13 15:38:26	2025-06-13 15:38:26
675	1	1	53	2094	2	-140	2025-03-21	\N	2025-06-13 15:38:27	2025-06-13 15:38:27
676	1	3	53	2101	2	-40	2025-03-21	\N	2025-06-13 15:38:32	2025-06-13 15:38:32
677	1	3	53	2102	2	-40	2025-03-21	\N	2025-06-13 15:38:34	2025-06-13 15:38:34
678	1	3	53	2108	2	-1000	2025-03-21	\N	2025-06-13 15:38:36	2025-06-13 15:38:36
679	1	11	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:33	2025-06-13 15:39:33
680	1	10	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:34	2025-06-13 15:39:34
681	1	8	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:34	2025-06-13 15:39:34
682	1	9	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:35	2025-06-13 15:39:35
683	1	2	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:35	2025-06-13 15:39:35
684	1	7	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:36	2025-06-13 15:39:36
685	1	3	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:36	2025-06-13 15:39:36
686	1	1	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:37	2025-06-13 15:39:37
687	1	4	54	\N	1	-1000	2025-04-18	\N	2025-06-13 15:39:38	2025-06-13 15:39:38
688	1	10	54	2109	2	-80	2025-04-18	\N	2025-06-13 15:40:42	2025-06-13 15:40:42
689	1	10	54	2110	2	-20	2025-04-18	\N	2025-06-13 15:40:43	2025-06-13 15:40:43
690	1	8	54	2117	2	-60	2025-04-18	\N	2025-06-13 15:40:55	2025-06-13 15:40:55
691	1	8	54	2118	2	-40	2025-04-18	\N	2025-06-13 15:40:56	2025-06-13 15:40:56
692	1	8	54	2119	2	-40	2025-04-18	\N	2025-06-13 15:40:57	2025-06-13 15:40:57
693	1	4	54	2125	2	-160	2025-04-18	\N	2025-06-13 15:41:06	2025-06-13 15:41:06
694	1	4	54	2126	2	-40	2025-04-18	\N	2025-06-13 15:41:06	2025-06-13 15:41:06
695	1	9	54	2133	2	-140	2025-04-18	\N	2025-06-13 15:41:12	2025-06-13 15:41:12
696	1	9	54	2134	2	-40	2025-04-18	\N	2025-06-13 15:41:13	2025-06-13 15:41:13
697	1	11	54	2141	2	-140	2025-04-18	\N	2025-06-13 15:41:24	2025-06-13 15:41:24
698	1	11	54	2142	2	-100	2025-04-18	\N	2025-06-13 15:41:24	2025-06-13 15:41:24
699	1	11	54	2145	2	-100	2025-04-18	\N	2025-06-13 15:41:26	2025-06-13 15:41:26
700	1	2	54	2149	2	-120	2025-04-18	\N	2025-06-13 15:41:38	2025-06-13 15:41:38
701	1	2	54	2150	2	-60	2025-04-18	\N	2025-06-13 15:41:39	2025-06-13 15:41:39
702	1	7	54	2157	2	-40	2025-04-18	\N	2025-06-13 15:41:45	2025-06-13 15:41:45
703	1	7	54	2158	2	-100	2025-04-18	\N	2025-06-13 15:41:47	2025-06-13 15:41:47
704	1	1	54	2165	2	-60	2025-04-18	\N	2025-06-13 15:41:52	2025-06-13 15:41:52
705	1	1	54	2166	2	-80	2025-04-18	\N	2025-06-13 15:41:53	2025-06-13 15:41:53
706	1	3	54	2174	2	-20	2025-04-18	\N	2025-06-13 15:41:59	2025-06-13 15:41:59
757	1	7	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:51	2025-08-06 10:41:51
758	1	2	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
759	1	3	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
760	1	8	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
761	1	9	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
762	1	10	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
763	1	11	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
764	1	1	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
765	1	4	56	\N	1	-1000	2025-07-11	\N	2025-08-06 10:41:52	2025-08-06 10:41:52
766	1	7	56	2245	2	-140	2025-07-11	\N	2025-08-06 10:43:21	2025-08-06 10:43:21
767	1	7	56	2246	2	-140	2025-07-11	\N	2025-08-06 10:43:22	2025-08-06 10:43:22
768	1	12	56	2253	2	-100	2025-07-11	\N	2025-08-06 10:43:51	2025-08-06 10:43:51
769	1	12	56	2254	2	-140	2025-07-11	\N	2025-08-06 10:43:52	2025-08-06 10:43:52
770	1	13	56	2261	2	-160	2025-07-11	\N	2025-08-06 10:44:09	2025-08-06 10:44:09
771	1	13	56	2262	2	-40	2025-07-11	\N	2025-08-06 10:44:09	2025-08-06 10:44:09
772	1	2	56	2269	2	-60	2025-07-11	\N	2025-08-06 10:44:26	2025-08-06 10:44:26
773	1	2	56	2270	2	-100	2025-07-11	\N	2025-08-06 10:44:26	2025-08-06 10:44:26
774	1	1	56	2277	2	-140	2025-07-11	\N	2025-08-06 10:44:42	2025-08-06 10:44:42
775	1	1	56	2278	2	-120	2025-07-11	\N	2025-08-06 10:44:42	2025-08-06 10:44:42
786	1	7	\N	\N	3	2260	2025-07-11	\N	2025-08-06 10:56:51	2025-08-06 10:56:51
787	1	2	\N	\N	3	1060	2025-07-11	\N	2025-08-06 10:57:40	2025-08-08 15:09:01
788	1	1	\N	\N	3	1180	2025-07-11	\N	2025-08-06 10:57:40	2025-08-08 15:09:14
822	1	9	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:38	2025-09-05 15:12:38
823	1	8	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:38	2025-09-05 15:12:38
824	1	11	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:39	2025-09-05 15:12:39
825	1	10	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:39	2025-09-05 15:12:39
826	1	1	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:40	2025-09-05 15:12:40
827	1	4	57	\N	1	-1000	2025-08-08	\N	2025-09-05 15:12:40	2025-09-05 15:12:40
828	1	10	57	2285	2	-20	2025-08-08	\N	2025-09-05 15:13:59	2025-09-05 15:13:59
829	1	10	57	2286	2	-60	2025-08-08	\N	2025-09-05 15:13:59	2025-09-05 15:13:59
830	1	8	57	2293	2	-140	2025-08-08	\N	2025-09-05 15:14:38	2025-09-05 15:14:38
831	1	8	57	2294	2	-160	2025-08-08	\N	2025-09-05 15:14:39	2025-09-05 15:14:39
832	1	8	57	2295	2	-20	2025-08-08	\N	2025-09-05 15:14:40	2025-09-05 15:14:40
833	1	8	57	2300	2	-500	2025-08-08	\N	2025-09-05 15:14:42	2025-09-05 15:14:42
834	1	9	57	2301	2	-100	2025-08-08	\N	2025-09-05 15:14:58	2025-09-05 15:14:58
835	1	9	57	2302	2	-120	2025-08-08	\N	2025-09-05 15:14:58	2025-09-05 15:14:58
836	1	11	57	2309	2	-60	2025-08-08	\N	2025-09-05 15:15:18	2025-09-05 15:15:18
837	1	11	57	2310	2	-60	2025-08-08	\N	2025-09-05 15:15:19	2025-09-05 15:15:19
838	1	11	57	2311	2	-20	2025-08-08	\N	2025-09-05 15:15:19	2025-09-05 15:15:19
839	1	2	57	2317	2	-80	2025-08-08	\N	2025-09-05 15:15:34	2025-09-05 15:15:34
840	1	2	57	2318	2	-40	2025-08-08	\N	2025-09-05 15:15:35	2025-09-05 15:15:35
841	1	1	57	2325	2	-80	2025-08-08	\N	2025-09-05 15:15:51	2025-09-05 15:15:51
842	1	1	57	2326	2	-120	2025-08-08	\N	2025-09-05 15:15:51	2025-09-05 15:15:51
843	1	1	57	2327	2	-20	2025-08-08	\N	2025-09-05 15:15:52	2025-09-05 15:15:52
844	1	7	57	2333	2	-100	2025-08-08	\N	2025-09-05 15:16:08	2025-09-05 15:16:08
845	1	7	57	2334	2	-140	2025-08-08	\N	2025-09-05 15:16:09	2025-09-05 15:16:09
846	1	3	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:24	2025-09-11 12:46:24
847	1	10	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:25	2025-09-11 12:46:25
848	1	8	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:26	2025-09-11 12:46:26
849	1	9	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:27	2025-09-11 12:46:27
850	1	11	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:28	2025-09-11 12:46:28
851	1	2	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:28	2025-09-11 12:46:28
852	1	7	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:29	2025-09-11 12:46:29
853	1	4	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:29	2025-09-11 12:46:29
854	1	1	58	\N	1	-1000	2025-09-05	\N	2025-09-11 12:46:29	2025-09-11 12:46:29
864	1	\N	\N	\N	5	-3000	2025-09-05	Kegelbahngebühr	2025-09-11 12:53:31	2025-09-11 12:54:22
865	1	10	58	2341	2	-40	2025-09-05	\N	2025-09-11 12:56:39	2025-09-11 12:56:39
866	1	10	58	2342	2	-100	2025-09-05	\N	2025-09-11 12:56:39	2025-09-11 12:56:39
867	1	8	58	2349	2	-240	2025-09-05	\N	2025-09-11 12:57:11	2025-09-11 12:57:11
855	1	2	\N	\N	3	1120	2025-09-05	\N	2025-09-11 12:50:24	2025-10-03 14:51:58
863	1	11	\N	\N	3	1140	2025-09-05	\N	2025-09-11 12:52:28	2025-10-03 14:52:45
862	1	10	\N	\N	3	1080	2025-09-05	\N	2025-09-11 12:52:28	2025-10-03 14:53:05
860	1	8	\N	\N	3	1820	2025-09-05	\N	2025-09-11 12:51:50	2025-10-03 14:53:46
861	1	9	\N	\N	3	1220	2025-09-05	\N	2025-09-11 12:51:50	2025-10-03 14:54:05
857	1	3	\N	\N	3	4100	2025-09-05	\N	2025-09-11 12:50:57	2025-10-03 14:54:48
858	1	4	\N	\N	3	3220	2025-09-05	\N	2025-09-11 12:50:57	2025-10-03 14:55:03
859	1	7	\N	\N	3	1240	2025-09-05	\N	2025-09-11 12:51:18	2025-10-03 15:02:36
868	1	8	58	2350	2	-20	2025-09-05	\N	2025-09-11 12:57:11	2025-09-11 12:57:11
869	1	8	58	2351	2	-80	2025-09-05	\N	2025-09-11 12:57:11	2025-09-11 12:57:11
870	1	8	58	2356	2	-500	2025-09-05	\N	2025-09-11 12:57:13	2025-09-11 12:57:13
871	1	4	58	2357	2	-220	2025-09-05	\N	2025-09-11 12:57:39	2025-09-11 12:57:39
872	1	4	58	2358	2	-120	2025-09-05	\N	2025-09-11 12:57:40	2025-09-11 12:57:40
873	1	9	58	2365	2	-240	2025-09-05	\N	2025-09-11 12:57:54	2025-09-11 12:57:54
874	1	9	58	2366	2	-120	2025-09-05	\N	2025-09-11 12:57:55	2025-09-11 12:57:55
875	1	11	58	2373	2	-60	2025-09-05	\N	2025-09-11 12:58:14	2025-09-11 12:58:14
876	1	11	58	2374	2	-80	2025-09-05	\N	2025-09-11 12:58:15	2025-09-11 12:58:15
877	1	7	58	2381	2	-220	2025-09-05	\N	2025-09-11 12:58:32	2025-09-11 12:58:32
878	1	7	58	2382	2	-80	2025-09-05	\N	2025-09-11 12:58:32	2025-09-11 12:58:32
879	1	1	58	2389	2	-120	2025-09-05	\N	2025-09-11 12:58:48	2025-09-11 12:58:48
880	1	1	58	2390	2	-100	2025-09-05	\N	2025-09-11 12:58:49	2025-09-11 12:58:49
881	1	1	58	2391	2	-20	2025-09-05	\N	2025-09-11 12:58:49	2025-09-11 12:58:49
882	1	3	58	2397	2	-100	2025-09-05	\N	2025-09-11 12:59:05	2025-09-11 12:59:05
883	1	3	58	2398	2	-20	2025-09-05	\N	2025-09-11 12:59:06	2025-09-11 12:59:06
856	1	1	\N	\N	3	1220	2025-09-05	\N	2025-09-11 12:50:24	2025-10-03 14:51:41
884	1	1	\N	\N	4	160	2025-09-05	\N	2025-10-03 14:52:26	2025-10-03 14:52:26
885	1	10	\N	\N	4	280	2025-09-05	\N	2025-10-03 14:53:24	2025-10-03 14:53:24
886	1	8	\N	\N	4	460	2025-09-05	\N	2025-10-03 14:54:29	2025-10-03 14:54:29
887	1	3	\N	\N	4	180	2025-09-05	\N	2025-10-03 14:55:28	2025-10-03 14:55:28
888	1	7	\N	\N	4	260	2025-09-05	\N	2025-10-03 15:02:56	2025-10-03 15:02:56
889	1	2	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:50	2025-10-31 16:51:50
890	1	11	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:50	2025-10-31 16:51:50
891	1	10	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:50	2025-10-31 16:51:50
892	1	8	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:51	2025-10-31 16:51:51
893	1	9	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:51	2025-10-31 16:51:51
894	1	3	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:52	2025-10-31 16:51:52
895	1	7	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:52	2025-10-31 16:51:52
896	1	1	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:53	2025-10-31 16:51:53
897	1	4	59	\N	1	-1000	2025-10-03	\N	2025-10-31 16:51:53	2025-10-31 16:51:53
898	1	2	\N	\N	3	2000	2025-10-31	\N	2025-11-28 15:23:05	2025-11-28 15:23:05
899	1	1	\N	\N	3	2240	2025-10-31	\N	2025-11-28 15:23:05	2025-11-28 15:23:05
900	1	2	\N	\N	4	260	2025-10-31	Auto Tip	2025-11-28 15:23:05	2025-11-28 15:23:05
901	1	8	\N	\N	3	2600	2025-10-31	\N	2025-11-28 15:24:12	2025-11-28 15:24:12
902	1	9	\N	\N	3	2600	2025-10-31	\N	2025-11-28 15:24:12	2025-11-28 15:24:12
903	1	11	\N	\N	3	2140	2025-10-31	\N	2025-11-28 15:25:00	2025-11-28 15:25:00
904	1	10	\N	\N	3	2140	2025-10-31	\N	2025-11-28 15:25:00	2025-11-28 15:25:00
905	1	11	\N	\N	4	220	2025-10-31	Auto Tip	2025-11-28 15:25:00	2025-11-28 15:25:00
906	1	3	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:25:57	2025-11-28 15:25:57
907	1	7	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:25:58	2025-11-28 15:25:58
908	1	2	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:25:58	2025-11-28 15:25:58
909	1	8	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:25:59	2025-11-28 15:25:59
910	1	9	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:25:59	2025-11-28 15:25:59
911	1	11	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:26:00	2025-11-28 15:26:00
912	1	10	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:26:00	2025-11-28 15:26:00
913	1	4	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:26:01	2025-11-28 15:26:01
914	1	1	60	\N	1	-1000	2025-10-31	\N	2025-11-28 15:26:02	2025-11-28 15:26:02
915	1	10	60	2405	2	-60	2025-10-31	\N	2025-11-28 15:27:41	2025-11-28 15:27:41
916	1	10	60	2406	2	-140	2025-10-31	\N	2025-11-28 15:27:42	2025-11-28 15:27:42
917	1	8	60	2413	2	-40	2025-10-31	\N	2025-11-28 15:28:07	2025-11-28 15:28:07
918	1	8	60	2414	2	-120	2025-10-31	\N	2025-11-28 15:28:08	2025-11-28 15:28:08
919	1	9	60	2421	2	-160	2025-10-31	\N	2025-11-28 15:28:39	2025-11-28 15:28:39
920	1	9	60	2422	2	-160	2025-10-31	\N	2025-11-28 15:28:39	2025-11-28 15:28:39
921	1	9	60	2425	2	-100	2025-10-31	\N	2025-11-28 15:28:39	2025-11-28 15:28:39
922	1	11	60	2429	2	-140	2025-10-31	\N	2025-11-28 15:29:07	2025-11-28 15:29:07
923	1	11	60	2430	2	-140	2025-10-31	\N	2025-11-28 15:29:07	2025-11-28 15:29:07
924	1	1	60	2437	2	-180	2025-10-31	\N	2025-11-28 15:29:27	2025-11-28 15:29:27
925	1	1	60	2438	2	-220	2025-10-31	\N	2025-11-28 15:29:28	2025-11-28 15:29:28
926	1	\N	\N	\N	5	3000	2025-10-31	Kegelbahn	2025-11-28 15:46:12	2025-11-28 15:46:12
927	1	2	\N	\N	3	1000	2025-11-28	\N	2026-01-23 15:23:23	2026-01-23 15:23:23
928	1	1	\N	\N	3	1400	2025-11-28	\N	2026-01-23 15:23:23	2026-01-23 15:23:23
929	1	2	\N	\N	4	100	2025-11-28	Auto Tip	2026-01-23 15:23:23	2026-01-23 15:23:23
936	1	8	\N	\N	3	1400	2025-11-28	\N	2026-01-23 15:27:06	2026-01-23 15:27:06
937	1	9	\N	\N	3	1180	2025-11-28	\N	2026-01-23 15:27:06	2026-01-23 15:27:06
938	1	8	\N	\N	4	120	2025-11-28	Auto Tip	2026-01-23 15:27:06	2026-01-23 15:27:06
939	1	10	\N	\N	3	1200	2025-11-28	\N	2026-01-23 15:27:55	2026-01-23 15:27:55
940	1	11	\N	\N	3	1280	2025-11-28	\N	2026-01-23 15:27:55	2026-01-23 15:27:55
941	1	10	\N	\N	4	20	2025-11-28	Auto Tip	2026-01-23 15:27:55	2026-01-23 15:27:55
942	1	7	\N	\N	3	3300	2025-11-28	\N	2026-01-23 15:28:17	2026-01-23 15:28:17
943	1	7	\N	\N	4	200	2025-11-28	\N	2026-01-23 15:28:17	2026-01-23 15:28:17
944	1	4	\N	\N	3	3340	2025-11-28	\N	2026-01-23 15:28:50	2026-01-23 15:28:50
945	1	3	\N	\N	3	3120	2025-11-28	\N	2026-01-23 15:29:29	2026-01-23 15:29:29
946	1	3	\N	\N	4	40	2025-11-28	\N	2026-01-23 15:29:29	2026-01-23 15:29:29
947	1	\N	\N	\N	5	3000	2025-11-28	Kegelbahn	2026-01-23 15:32:54	2026-01-23 15:32:54
948	1	2	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
949	1	8	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
950	1	9	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
951	1	10	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
952	1	11	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
953	1	7	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
954	1	3	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
955	1	1	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
956	1	4	61	\N	1	-1000	2025-11-28	\N	2026-01-23 15:33:46	2026-01-23 15:33:46
957	1	10	61	2445	2	-180	2025-11-28	\N	2026-01-23 15:35:44	2026-01-23 15:35:44
958	1	10	61	2446	2	-80	2025-11-28	\N	2026-01-23 15:35:44	2026-01-23 15:35:44
959	1	8	61	2453	2	-140	2025-11-28	\N	2026-01-23 15:36:09	2026-01-23 15:36:09
960	1	8	61	2454	2	-60	2025-11-28	\N	2026-01-23 15:36:09	2026-01-23 15:36:09
961	1	4	61	2461	2	-300	2025-11-28	\N	2026-01-23 15:36:24	2026-01-23 15:36:24
962	1	4	61	2462	2	-40	2025-11-28	\N	2026-01-23 15:36:25	2026-01-23 15:36:25
963	1	9	61	2469	2	-160	2025-11-28	\N	2026-01-23 15:36:42	2026-01-23 15:36:42
964	1	9	61	2470	2	-80	2025-11-28	\N	2026-01-23 15:36:43	2026-01-23 15:36:43
965	1	2	61	2477	2	-240	2025-11-28	\N	2026-01-23 15:37:05	2026-01-23 15:37:05
966	1	2	61	2478	2	-60	2025-11-28	\N	2026-01-23 15:37:06	2026-01-23 15:37:06
967	1	7	61	2485	2	-140	2025-11-28	\N	2026-01-23 15:37:24	2026-01-23 15:37:24
968	1	7	61	2486	2	-120	2025-11-28	\N	2026-01-23 15:37:25	2026-01-23 15:37:25
969	1	1	61	2493	2	-160	2025-11-28	\N	2026-01-23 15:37:47	2026-01-23 15:37:47
970	1	1	61	2494	2	-120	2025-11-28	\N	2026-01-23 15:37:48	2026-01-23 15:37:48
971	1	1	61	2495	2	-40	2025-11-28	\N	2026-01-23 15:37:49	2026-01-23 15:37:49
972	1	3	61	2501	2	-140	2025-11-28	\N	2026-01-23 15:38:00	2026-01-23 15:38:00
973	1	3	61	2502	2	-20	2025-11-28	\N	2026-01-23 15:38:01	2026-01-23 15:38:01
\.


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

COPY public.users (id, name, email, email_verified_at, password, remember_token, created_at, updated_at) FROM stdin;
2	Michael	m.ludwig182@icloud.com	2025-03-08 13:28:36	$2y$12$cCCLCcqXNtMtPsdVKYSl/eb1AAk6d1TBdBnmkoDUC4Uh34UZBt.GC	\N	2025-03-08 13:27:28	2025-03-08 13:28:36
4	A.Stranger	a.stranger@gmx.net	2025-03-14 06:37:39	$2y$12$KLxvo/DtkEIOTRWAzEGLd.horjx20n9r6YXwgWB6WIR1C.vI5DWIe	\N	2025-03-14 06:37:25	2025-03-14 06:37:39
1	Pascal	pascal@schnurbus.de	2025-03-08 11:34:13	$2y$12$1evzjPEbb8ZPBYXrtZF3juxbu7YNxEMi2m4QuahvvpW2h..KRTFY2	WVsdLhJhT78kiClyK2E1UeewYWvgxrpUb2rFEuz7V5qp9Oi5SPeYnv2dn0WG	2025-03-08 11:33:50	2025-03-08 11:34:13
3	Nicole	nicole@schnurbus.de	2025-03-09 16:33:35	$2y$12$PkhNTdYYEbPNwBF9yR4me.4cFnEHIn65HN9qH9k1FRUvOdKo7T0fe	7xCckYRjmnQUipfJOpaYZhgtrKLwnsFojmzdecA9HFSOvB2UbkCF561v9sEs	2025-03-09 16:33:14	2025-03-09 16:33:35
\.


--
-- Name: club_settings_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.club_settings_id_seq', 1, false);


--
-- Name: clubs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.clubs_id_seq', 6, true);


--
-- Name: competition_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.competition_entries_id_seq', 373, true);


--
-- Name: competition_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.competition_types_id_seq', 4, true);


--
-- Name: dashboard_layouts_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.dashboard_layouts_id_seq', 8, true);


--
-- Name: failed_jobs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.failed_jobs_id_seq', 1, false);


--
-- Name: fee_entries_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.fee_entries_id_seq', 2508, true);


--
-- Name: fee_type_versions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.fee_type_versions_id_seq', 52, true);


--
-- Name: fee_types_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.fee_types_id_seq', 25, true);


--
-- Name: jobs_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.jobs_id_seq', 1, false);


--
-- Name: matchdays_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.matchdays_id_seq', 61, true);


--
-- Name: migrations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.migrations_id_seq', 27, true);


--
-- Name: permissions_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.permissions_id_seq1', 61, true);


--
-- Name: personal_access_tokens_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.personal_access_tokens_id_seq', 1, false);


--
-- Name: player_invitations_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.player_invitations_id_seq', 5, true);


--
-- Name: players_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.players_id_seq', 37, true);


--
-- Name: roles_id_seq1; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.roles_id_seq1', 14, true);


--
-- Name: transactions_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.transactions_id_seq', 973, true);


--
-- Name: users_id_seq; Type: SEQUENCE SET; Schema: public; Owner: -
--

SELECT pg_catalog.setval('public.users_id_seq', 4, true);


--
-- Name: cache_locks cache_locks_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cache_locks
    ADD CONSTRAINT cache_locks_pkey PRIMARY KEY (key);


--
-- Name: cache cache_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.cache
    ADD CONSTRAINT cache_pkey PRIMARY KEY (key);


--
-- Name: club_settings club_settings_club_id_name_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.club_settings
    ADD CONSTRAINT club_settings_club_id_name_unique UNIQUE (club_id, name);


--
-- Name: club_settings club_settings_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.club_settings
    ADD CONSTRAINT club_settings_pkey PRIMARY KEY (id);


--
-- Name: clubs clubs_name_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_name_unique UNIQUE (name);


--
-- Name: clubs clubs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_pkey PRIMARY KEY (id);


--
-- Name: competition_entries competition_entries_matchday_id_player_id_competition_type_id_u; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_entries
    ADD CONSTRAINT competition_entries_matchday_id_player_id_competition_type_id_u UNIQUE (matchday_id, player_id, competition_type_id);


--
-- Name: competition_entries competition_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_entries
    ADD CONSTRAINT competition_entries_pkey PRIMARY KEY (id);


--
-- Name: competition_types competition_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_types
    ADD CONSTRAINT competition_types_pkey PRIMARY KEY (id);


--
-- Name: dashboard_layouts dashboard_layouts_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dashboard_layouts
    ADD CONSTRAINT dashboard_layouts_pkey PRIMARY KEY (id);


--
-- Name: failed_jobs failed_jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_jobs
    ADD CONSTRAINT failed_jobs_pkey PRIMARY KEY (id);


--
-- Name: failed_jobs failed_jobs_uuid_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.failed_jobs
    ADD CONSTRAINT failed_jobs_uuid_unique UNIQUE (uuid);


--
-- Name: fee_entries fee_entries_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_entries
    ADD CONSTRAINT fee_entries_pkey PRIMARY KEY (id);


--
-- Name: fee_type_version_matchday fee_type_version_matchday_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_type_version_matchday
    ADD CONSTRAINT fee_type_version_matchday_pkey PRIMARY KEY (fee_type_version_id, matchday_id);


--
-- Name: fee_type_versions fee_type_versions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_type_versions
    ADD CONSTRAINT fee_type_versions_pkey PRIMARY KEY (id);


--
-- Name: fee_types fee_types_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_types
    ADD CONSTRAINT fee_types_pkey PRIMARY KEY (id);


--
-- Name: job_batches job_batches_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.job_batches
    ADD CONSTRAINT job_batches_pkey PRIMARY KEY (id);


--
-- Name: jobs jobs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.jobs
    ADD CONSTRAINT jobs_pkey PRIMARY KEY (id);


--
-- Name: matchday_player matchday_player_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matchday_player
    ADD CONSTRAINT matchday_player_pkey PRIMARY KEY (matchday_id, player_id);


--
-- Name: matchdays matchdays_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matchdays
    ADD CONSTRAINT matchdays_pkey PRIMARY KEY (id);


--
-- Name: migrations migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.migrations
    ADD CONSTRAINT migrations_pkey PRIMARY KEY (id);


--
-- Name: model_has_permissions model_has_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_has_permissions
    ADD CONSTRAINT model_has_permissions_pkey PRIMARY KEY (club_id, permission_id, model_id, model_type);


--
-- Name: model_has_roles model_has_roles_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_has_roles
    ADD CONSTRAINT model_has_roles_pkey PRIMARY KEY (club_id, role_id, model_id, model_type);


--
-- Name: password_reset_tokens password_reset_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.password_reset_tokens
    ADD CONSTRAINT password_reset_tokens_pkey PRIMARY KEY (email);


--
-- Name: permissions permissions_name_guard_name_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_name_guard_name_unique UNIQUE (name, guard_name);


--
-- Name: permissions permissions_pkey1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.permissions
    ADD CONSTRAINT permissions_pkey1 PRIMARY KEY (id);


--
-- Name: personal_access_tokens personal_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.personal_access_tokens
    ADD CONSTRAINT personal_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: personal_access_tokens personal_access_tokens_token_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.personal_access_tokens
    ADD CONSTRAINT personal_access_tokens_token_unique UNIQUE (token);


--
-- Name: player_invitations player_invitations_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.player_invitations
    ADD CONSTRAINT player_invitations_pkey PRIMARY KEY (id);


--
-- Name: players players_name_club_id_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT players_name_club_id_unique UNIQUE (name, club_id);


--
-- Name: players players_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT players_pkey PRIMARY KEY (id);


--
-- Name: role_has_permissions role_has_permissions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_has_permissions
    ADD CONSTRAINT role_has_permissions_pkey PRIMARY KEY (permission_id, role_id);


--
-- Name: roles roles_club_id_name_guard_name_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_club_id_name_guard_name_unique UNIQUE (club_id, name, guard_name);


--
-- Name: roles roles_pkey1; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.roles
    ADD CONSTRAINT roles_pkey1 PRIMARY KEY (id);


--
-- Name: sessions sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.sessions
    ADD CONSTRAINT sessions_pkey PRIMARY KEY (id);


--
-- Name: transactions transactions_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_pkey PRIMARY KEY (id);


--
-- Name: users users_email_unique; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_unique UNIQUE (email);


--
-- Name: users users_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);


--
-- Name: club_settings_club_id_name_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX club_settings_club_id_name_index ON public.club_settings USING btree (club_id, name);


--
-- Name: competition_entries_matchday_id_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX competition_entries_matchday_id_index ON public.competition_entries USING btree (matchday_id);


--
-- Name: competition_types_club_id_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX competition_types_club_id_index ON public.competition_types USING btree (club_id);


--
-- Name: dashboard_layouts_user_id_club_id_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX dashboard_layouts_user_id_club_id_index ON public.dashboard_layouts USING btree (user_id, club_id);


--
-- Name: jobs_queue_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX jobs_queue_index ON public.jobs USING btree (queue);


--
-- Name: model_has_permissions_model_id_model_type_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX model_has_permissions_model_id_model_type_index ON public.model_has_permissions USING btree (model_id, model_type);


--
-- Name: model_has_permissions_team_foreign_key_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX model_has_permissions_team_foreign_key_index ON public.model_has_permissions USING btree (club_id);


--
-- Name: model_has_roles_model_id_model_type_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX model_has_roles_model_id_model_type_index ON public.model_has_roles USING btree (model_id, model_type);


--
-- Name: model_has_roles_team_foreign_key_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX model_has_roles_team_foreign_key_index ON public.model_has_roles USING btree (club_id);


--
-- Name: personal_access_tokens_tokenable_type_tokenable_id_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX personal_access_tokens_tokenable_type_tokenable_id_index ON public.personal_access_tokens USING btree (tokenable_type, tokenable_id);


--
-- Name: roles_team_foreign_key_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX roles_team_foreign_key_index ON public.roles USING btree (club_id);


--
-- Name: sessions_last_activity_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sessions_last_activity_index ON public.sessions USING btree (last_activity);


--
-- Name: sessions_user_id_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX sessions_user_id_index ON public.sessions USING btree (user_id);


--
-- Name: transactions_club_id_type_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_club_id_type_index ON public.transactions USING btree (club_id, type);


--
-- Name: transactions_player_id_type_index; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX transactions_player_id_type_index ON public.transactions USING btree (player_id, type);


--
-- Name: competition_entries 1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_entries
    ADD CONSTRAINT "1" FOREIGN KEY (matchday_id) REFERENCES public.matchdays(id);


--
-- Name: matchdays 1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matchdays
    ADD CONSTRAINT "1" FOREIGN KEY (club_id) REFERENCES public.clubs(id);


--
-- Name: players 1; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT "1" FOREIGN KEY (club_id) REFERENCES public.clubs(id);


--
-- Name: club_settings club_settings_club_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.club_settings
    ADD CONSTRAINT club_settings_club_id_foreign FOREIGN KEY (club_id) REFERENCES public.clubs(id) ON DELETE CASCADE;


--
-- Name: clubs clubs_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.clubs
    ADD CONSTRAINT clubs_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: competition_entries competition_entries_competition_type_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_entries
    ADD CONSTRAINT competition_entries_competition_type_id_foreign FOREIGN KEY (competition_type_id) REFERENCES public.competition_types(id);


--
-- Name: competition_entries competition_entries_player_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_entries
    ADD CONSTRAINT competition_entries_player_id_foreign FOREIGN KEY (player_id) REFERENCES public.players(id);


--
-- Name: competition_types competition_types_club_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.competition_types
    ADD CONSTRAINT competition_types_club_id_foreign FOREIGN KEY (club_id) REFERENCES public.clubs(id) ON DELETE CASCADE;


--
-- Name: dashboard_layouts dashboard_layouts_club_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dashboard_layouts
    ADD CONSTRAINT dashboard_layouts_club_id_foreign FOREIGN KEY (club_id) REFERENCES public.clubs(id) ON DELETE CASCADE;


--
-- Name: dashboard_layouts dashboard_layouts_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.dashboard_layouts
    ADD CONSTRAINT dashboard_layouts_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id) ON DELETE CASCADE;


--
-- Name: fee_entries fee_entries_fee_type_version_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_entries
    ADD CONSTRAINT fee_entries_fee_type_version_id_foreign FOREIGN KEY (fee_type_version_id) REFERENCES public.fee_type_versions(id) ON DELETE CASCADE;


--
-- Name: fee_entries fee_entries_matchday_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_entries
    ADD CONSTRAINT fee_entries_matchday_id_foreign FOREIGN KEY (matchday_id) REFERENCES public.matchdays(id) ON DELETE CASCADE;


--
-- Name: fee_entries fee_entries_player_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_entries
    ADD CONSTRAINT fee_entries_player_id_foreign FOREIGN KEY (player_id) REFERENCES public.players(id) ON DELETE CASCADE;


--
-- Name: fee_type_version_matchday fee_type_version_matchday_fee_type_version_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_type_version_matchday
    ADD CONSTRAINT fee_type_version_matchday_fee_type_version_id_foreign FOREIGN KEY (fee_type_version_id) REFERENCES public.fee_type_versions(id) ON DELETE CASCADE;


--
-- Name: fee_type_version_matchday fee_type_version_matchday_matchday_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_type_version_matchday
    ADD CONSTRAINT fee_type_version_matchday_matchday_id_foreign FOREIGN KEY (matchday_id) REFERENCES public.matchdays(id) ON DELETE CASCADE;


--
-- Name: fee_type_versions fee_type_versions_fee_type_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_type_versions
    ADD CONSTRAINT fee_type_versions_fee_type_id_foreign FOREIGN KEY (fee_type_id) REFERENCES public.fee_types(id) ON DELETE CASCADE;


--
-- Name: fee_types fee_types_club_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.fee_types
    ADD CONSTRAINT fee_types_club_id_foreign FOREIGN KEY (club_id) REFERENCES public.clubs(id) ON DELETE CASCADE;


--
-- Name: matchday_player matchday_player_matchday_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matchday_player
    ADD CONSTRAINT matchday_player_matchday_id_foreign FOREIGN KEY (matchday_id) REFERENCES public.matchdays(id) ON DELETE CASCADE;


--
-- Name: matchday_player matchday_player_player_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.matchday_player
    ADD CONSTRAINT matchday_player_player_id_foreign FOREIGN KEY (player_id) REFERENCES public.players(id) ON DELETE CASCADE;


--
-- Name: model_has_permissions model_has_permissions_permission_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_has_permissions
    ADD CONSTRAINT model_has_permissions_permission_id_foreign FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- Name: model_has_roles model_has_roles_role_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.model_has_roles
    ADD CONSTRAINT model_has_roles_role_id_foreign FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: player_invitations player_invitations_player_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.player_invitations
    ADD CONSTRAINT player_invitations_player_id_foreign FOREIGN KEY (player_id) REFERENCES public.players(id) ON DELETE CASCADE;


--
-- Name: players players_role_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT players_role_id_foreign FOREIGN KEY (role_id) REFERENCES public.roles(id);


--
-- Name: players players_user_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.players
    ADD CONSTRAINT players_user_id_foreign FOREIGN KEY (user_id) REFERENCES public.users(id);


--
-- Name: role_has_permissions role_has_permissions_permission_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_has_permissions
    ADD CONSTRAINT role_has_permissions_permission_id_foreign FOREIGN KEY (permission_id) REFERENCES public.permissions(id) ON DELETE CASCADE;


--
-- Name: role_has_permissions role_has_permissions_role_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.role_has_permissions
    ADD CONSTRAINT role_has_permissions_role_id_foreign FOREIGN KEY (role_id) REFERENCES public.roles(id) ON DELETE CASCADE;


--
-- Name: transactions transactions_club_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_club_id_foreign FOREIGN KEY (club_id) REFERENCES public.clubs(id) ON DELETE CASCADE;


--
-- Name: transactions transactions_fee_entry_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_fee_entry_id_foreign FOREIGN KEY (fee_entry_id) REFERENCES public.fee_entries(id);


--
-- Name: transactions transactions_matchday_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_matchday_id_foreign FOREIGN KEY (matchday_id) REFERENCES public.matchdays(id);


--
-- Name: transactions transactions_player_id_foreign; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.transactions
    ADD CONSTRAINT transactions_player_id_foreign FOREIGN KEY (player_id) REFERENCES public.players(id);


--
-- PostgreSQL database dump complete
--

\unrestrict AFqcUbbLzJDNkRQw0eYdMOR8cbxDNvzA6BVVLUsm6gGgiIIsBBL6r2qeDYP5wCA

