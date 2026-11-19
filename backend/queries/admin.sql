-- Instanz-Administration (read-only Betreiber-Übersicht, KEG-7).
-- Bewusst KEINE Fachdaten der Klubs (Salden, Transaktionen, Strafen einzelner
-- Spieler); höchstens aggregierte Anzahlen.
--
-- Gemeinsame Definitionen (in allen Queries gleich halten):
--   verwaist      = Benutzer ohne eigenen Klub und ohne Spieler-Verknüpfung,
--                   registriert vor mehr als 30 Tagen
--   Klub-Aktivität = MAX(game_days.date, transactions.created_at)
--   aktiver Klub   = Aktivität in den letzten 90 Tagen
--   inaktiver Klub = keine Aktivität seit 180 Tagen (ohne Aktivität: angelegt vor 180 Tagen)

-- name: GetAdminStats :one
SELECT
    (SELECT COUNT(*) FROM users)::bigint AS users_total,
    (SELECT COUNT(*) FROM users u WHERE u.created_at >= now() - interval '7 days')::bigint AS users_new_7d,
    (SELECT COUNT(*) FROM users u WHERE u.created_at >= now() - interval '30 days')::bigint AS users_new_30d,
    (SELECT COUNT(*) FROM users u WHERE u.last_login_at >= now() - interval '30 days')::bigint AS users_active_30d,
    (SELECT COUNT(*) FROM users u WHERE u.is_instance_admin)::bigint AS users_instance_admins,
    (SELECT COUNT(*) FROM users u
        WHERE u.created_at < now() - interval '30 days'
          AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.user_id = u.id)
          AND NOT EXISTS (SELECT 1 FROM players p WHERE p.user_id = u.id))::bigint AS users_orphaned,
    (SELECT COUNT(*) FROM clubs)::bigint AS clubs_total,
    (SELECT COUNT(*) FROM clubs c
        WHERE EXISTS (SELECT 1 FROM game_days g WHERE g.club_id = c.id AND g.date >= (now() - interval '90 days')::date)
           OR EXISTS (SELECT 1 FROM transactions t WHERE t.club_id = c.id AND t.created_at >= now() - interval '90 days'))::bigint AS clubs_active_90d,
    (SELECT COUNT(*) FROM clubs c
        WHERE c.created_at < now() - interval '180 days'
          AND NOT EXISTS (SELECT 1 FROM game_days g WHERE g.club_id = c.id AND g.date >= (now() - interval '180 days')::date)
          AND NOT EXISTS (SELECT 1 FROM transactions t WHERE t.club_id = c.id AND t.created_at >= now() - interval '180 days'))::bigint AS clubs_inactive_180d,
    (SELECT COUNT(*) FROM players)::bigint AS players_total,
    (SELECT COUNT(*) FROM players p WHERE p.user_id IS NOT NULL)::bigint AS players_linked,
    (SELECT COUNT(*) FROM player_invitations i WHERE i.accepted_at IS NULL AND i.expires_at > now())::bigint AS invitations_open,
    (SELECT COUNT(*) FROM api_tokens t WHERE t.expires_at IS NULL OR t.expires_at > now())::bigint AS api_tokens_active,
    (SELECT COUNT(*) FROM api_tokens t WHERE t.expires_at IS NULL)::bigint AS api_tokens_without_expiry,
    (SELECT COUNT(*) FROM api_tokens t
        WHERE (t.expires_at IS NULL OR t.expires_at > now())
          AND COALESCE(t.last_used_at, t.created_at) < now() - interval '90 days')::bigint AS api_tokens_unused_90d,
    (SELECT COUNT(*) FROM webauthn_credentials)::bigint AS passkeys_total;

-- name: ListAdminRegistrationSpikes :many
-- Tage der letzten 30 Tage mit mehr als min_count Registrierungen (UTC).
SELECT
    (u.created_at AT TIME ZONE 'UTC')::date AS day,
    COUNT(*)::bigint AS registrations
FROM users u
WHERE u.created_at >= now() - interval '30 days'
GROUP BY day
HAVING COUNT(*) > sqlc.arg(min_count)::bigint
ORDER BY day DESC;

-- name: ListAdminUsers :many
SELECT
    u.id,
    u.email,
    u.is_instance_admin,
    u.created_at,
    u.last_login_at,
    (SELECT COUNT(*) FROM clubs c WHERE c.user_id = u.id)::bigint AS owned_club_count,
    (SELECT COUNT(DISTINCT p.club_id) FROM players p WHERE p.user_id = u.id)::bigint AS player_club_count,
    (SELECT COUNT(*) FROM webauthn_credentials w WHERE w.user_id = u.id)::bigint AS passkey_count,
    (SELECT COUNT(*) FROM api_tokens t WHERE t.user_id = u.id)::bigint AS api_token_count,
    (u.created_at < now() - interval '30 days'
        AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.user_id = u.id)
        AND NOT EXISTS (SELECT 1 FROM players p WHERE p.user_id = u.id))::boolean AS orphaned
FROM users u
WHERE (sqlc.arg(search)::text = '' OR u.email ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (NOT sqlc.arg(only_admins)::boolean OR u.is_instance_admin)
  AND (NOT sqlc.arg(only_orphaned)::boolean OR (
        u.created_at < now() - interval '30 days'
        AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.user_id = u.id)
        AND NOT EXISTS (SELECT 1 FROM players p WHERE p.user_id = u.id)))
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'email_asc' THEN u.email END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'email_desc' THEN u.email END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'created_at_asc' THEN u.created_at END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'created_at_desc' THEN u.created_at END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'last_login_at_asc' THEN u.last_login_at END ASC NULLS FIRST,
    CASE WHEN sqlc.arg(sort)::text = 'last_login_at_desc' THEN u.last_login_at END DESC NULLS LAST,
    u.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: CountAdminUsers :one
SELECT COUNT(*)::bigint
FROM users u
WHERE (sqlc.arg(search)::text = '' OR u.email ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (NOT sqlc.arg(only_admins)::boolean OR u.is_instance_admin)
  AND (NOT sqlc.arg(only_orphaned)::boolean OR (
        u.created_at < now() - interval '30 days'
        AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.user_id = u.id)
        AND NOT EXISTS (SELECT 1 FROM players p WHERE p.user_id = u.id)));

-- name: GetAdminUser :one
SELECT
    u.id,
    u.email,
    u.is_instance_admin,
    u.created_at,
    u.updated_at,
    u.password_changed_at,
    u.last_login_at,
    (u.created_at < now() - interval '30 days'
        AND NOT EXISTS (SELECT 1 FROM clubs c WHERE c.user_id = u.id)
        AND NOT EXISTS (SELECT 1 FROM players p WHERE p.user_id = u.id))::boolean AS orphaned
FROM users u
WHERE u.id = $1;

-- name: ListAdminUserOwnedClubs :many
SELECT
    c.id,
    c.name,
    c.created_at,
    (SELECT COUNT(*) FROM players p WHERE p.club_id = c.id)::bigint AS player_count
FROM clubs c
WHERE c.user_id = $1
ORDER BY c.name, c.id;

-- name: ListAdminUserPlayers :many
SELECT
    p.id,
    p.name,
    p.inactive,
    p.club_id,
    c.name AS club_name,
    (c.user_id = p.user_id)::boolean AS is_owner,
    p.role_id,
    r.name AS role_name
FROM players p
JOIN clubs c ON c.id = p.club_id
LEFT JOIN roles r ON r.id = p.role_id
WHERE p.user_id = $1
ORDER BY c.name, p.name, p.id;

-- name: ListAdminUserAPITokens :many
-- Nie den Hash ausliefern.
SELECT id, name, token_prefix, created_at, expires_at, last_used_at
FROM api_tokens
WHERE user_id = $1
ORDER BY created_at DESC, id;

-- name: ListAdminUserPasskeys :many
SELECT id, name, created_at, last_used_at
FROM webauthn_credentials
WHERE user_id = $1
ORDER BY created_at DESC, id;

-- name: ListAdminClubs :many
SELECT
    c.id,
    c.name,
    c.created_at,
    c.user_id AS owner_id,
    u.email AS owner_email,
    (SELECT COUNT(*) FROM players p WHERE p.club_id = c.id)::bigint AS player_count,
    (SELECT COUNT(*) FROM players p WHERE p.club_id = c.id AND p.user_id IS NOT NULL)::bigint AS linked_player_count,
    (SELECT COUNT(*) FROM game_days g WHERE g.club_id = c.id)::bigint AS game_day_count,
    a.last_activity_at
FROM clubs c
JOIN users u ON u.id = c.user_id
CROSS JOIN LATERAL (
    SELECT GREATEST(
        (SELECT MAX(g.date)::timestamptz FROM game_days g WHERE g.club_id = c.id),
        (SELECT MAX(t.created_at) FROM transactions t WHERE t.club_id = c.id)
    ) AS last_activity_at
) a
WHERE (sqlc.arg(search)::text = ''
        OR c.name ILIKE '%' || sqlc.arg(search)::text || '%'
        OR u.email ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (NOT sqlc.arg(only_inactive)::boolean
        OR COALESCE(a.last_activity_at, c.created_at) < now() - interval '180 days')
ORDER BY
    CASE WHEN sqlc.arg(sort)::text = 'name_asc' THEN c.name END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'name_desc' THEN c.name END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'created_at_asc' THEN c.created_at END ASC,
    CASE WHEN sqlc.arg(sort)::text = 'created_at_desc' THEN c.created_at END DESC,
    CASE WHEN sqlc.arg(sort)::text = 'last_activity_at_asc' THEN a.last_activity_at END ASC NULLS FIRST,
    CASE WHEN sqlc.arg(sort)::text = 'last_activity_at_desc' THEN a.last_activity_at END DESC NULLS LAST,
    c.id
LIMIT sqlc.arg(page_limit)::int OFFSET sqlc.arg(page_offset)::int;

-- name: CountAdminClubs :one
SELECT COUNT(*)::bigint
FROM clubs c
JOIN users u ON u.id = c.user_id
CROSS JOIN LATERAL (
    SELECT GREATEST(
        (SELECT MAX(g.date)::timestamptz FROM game_days g WHERE g.club_id = c.id),
        (SELECT MAX(t.created_at) FROM transactions t WHERE t.club_id = c.id)
    ) AS last_activity_at
) a
WHERE (sqlc.arg(search)::text = ''
        OR c.name ILIKE '%' || sqlc.arg(search)::text || '%'
        OR u.email ILIKE '%' || sqlc.arg(search)::text || '%')
  AND (NOT sqlc.arg(only_inactive)::boolean
        OR COALESCE(a.last_activity_at, c.created_at) < now() - interval '180 days');

-- name: GetAdminClub :one
SELECT
    c.id,
    c.name,
    c.created_at,
    c.user_id AS owner_id,
    u.email AS owner_email,
    (SELECT COUNT(*) FROM players p WHERE p.club_id = c.id)::bigint AS player_count,
    (SELECT COUNT(*) FROM players p WHERE p.club_id = c.id AND p.user_id IS NOT NULL)::bigint AS linked_player_count,
    (SELECT COUNT(*) FROM game_days g WHERE g.club_id = c.id)::bigint AS game_day_count,
    (SELECT COUNT(*) FROM transactions t WHERE t.club_id = c.id)::bigint AS transaction_count,
    a.last_activity_at
FROM clubs c
JOIN users u ON u.id = c.user_id
CROSS JOIN LATERAL (
    SELECT GREATEST(
        (SELECT MAX(g.date)::timestamptz FROM game_days g WHERE g.club_id = c.id),
        (SELECT MAX(t.created_at) FROM transactions t WHERE t.club_id = c.id)
    ) AS last_activity_at
) a
WHERE c.id = $1;

-- name: ListAdminClubPlayers :many
SELECT
    p.id,
    p.name,
    p.inactive,
    p.created_at,
    p.user_id,
    u.email AS user_email,
    p.role_id,
    r.name AS role_name
FROM players p
LEFT JOIN users u ON u.id = p.user_id
LEFT JOIN roles r ON r.id = p.role_id
WHERE p.club_id = $1
ORDER BY p.name, p.id;

-- name: ListAdminClubOpenInvitations :many
SELECT
    i.id,
    i.player_id,
    p.name AS player_name,
    i.email,
    i.created_at,
    i.expires_at
FROM player_invitations i
JOIN players p ON p.id = i.player_id
WHERE p.club_id = $1
  AND i.accepted_at IS NULL
  AND i.expires_at > now()
ORDER BY i.created_at DESC, i.id;

-- name: ListAdminClubRoles :many
SELECT
    r.id,
    r.name,
    (SELECT COUNT(*) FROM players p WHERE p.role_id = r.id)::bigint AS player_count
FROM roles r
WHERE r.club_id = $1
ORDER BY r.name, r.id;

-- name: ListAdminMemberships :many
-- CSV-Export: eine Zeile pro Spieler-Verknüpfung, dazu eine Zeile pro Klub,
-- in dem der Owner keinen eigenen Spieler hat. Fehlende Werte sind ''.
SELECT
    m.user_email::text AS user_email,
    m.club_name::text AS club_name,
    m.is_owner::boolean AS is_owner,
    m.player_name::text AS player_name,
    m.role_name::text AS role_name
FROM (
    SELECT
        u.email AS user_email,
        c.name AS club_name,
        (c.user_id = u.id) AS is_owner,
        p.name AS player_name,
        COALESCE(r.name, '') AS role_name
    FROM players p
    JOIN users u ON u.id = p.user_id
    JOIN clubs c ON c.id = p.club_id
    LEFT JOIN roles r ON r.id = p.role_id
    UNION ALL
    SELECT
        u.email AS user_email,
        c.name AS club_name,
        true AS is_owner,
        '' AS player_name,
        '' AS role_name
    FROM clubs c
    JOIN users u ON u.id = c.user_id
    WHERE NOT EXISTS (SELECT 1 FROM players p WHERE p.club_id = c.id AND p.user_id = c.user_id)
) m
ORDER BY m.user_email, m.club_name, m.player_name;

-- name: SetInstanceAdminByEmail :one
-- Für CLI und INSTANCE_ADMIN_EMAILS. Liefert keine Zeile, wenn das Konto fehlt.
UPDATE users
SET is_instance_admin = sqlc.arg(is_instance_admin)::boolean
WHERE email = lower(sqlc.arg(email)::text)
RETURNING id, email, is_instance_admin;

-- name: ListInstanceAdmins :many
SELECT id, email, created_at, last_login_at
FROM users
WHERE is_instance_admin
ORDER BY email;
