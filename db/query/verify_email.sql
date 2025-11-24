-- name: CreateVerifyEmail :one
INSERT INTO "Verify_Emails" (
    username,
    email,
    secret_code
    ) VALUES (
    $1, $2, $3
    ) 
    RETURNING *;

-- name: UpdateVerifyEmail :one
UPDATE "Verify_Emails"
SET
    is_used = true
WHERE
    email_id = @email_id
    AND secret_code = @secret_code
    AND is_used = FALSE
    AND expired_at > now()
    RETURNING *;