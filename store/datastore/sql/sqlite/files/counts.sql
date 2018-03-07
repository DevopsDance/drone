-- name: count-users

SELECT count(1)
FROM users

-- name: count-repos

SELECT count(1)
FROM repos
WHERE repo_active = 1

-- name: count-builds

SELECT count(1)
FROM builds

-- name: count-builds-failed

SELECT count(1)
FROM builds
WHERE build_status == "failure"

-- name: count-builds-succeded

SELECT count(1)
FROM builds
WHERE build_status == "success"
