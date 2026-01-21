-- name: GetVerifiedYoutubeAnalysisByRaidID :many
SELECT id, video_id, raid_id, analysis_result, analysis_type, version, is_verified, created_at, updated_at
FROM batorment_v3.youtube_analysis
WHERE raid_id = $1
  AND is_verified = true
  AND deleted_at IS NULL
ORDER BY created_at DESC;
