# Notify to Slack

Notify message to slack from gcs storage file.

## Deploy

```
$ gcloud functions deploy notify_to_slack \
    --entry-point NotifyToSlack \
    --runtime go111 \
    --set-env-vars 'WEBHOOK_URL=...' \
    --trigger-bucket <bucket> \
    --project <your_gcp_project_id> \
    --region asia-northeast1
```
