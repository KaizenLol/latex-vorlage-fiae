resource "aws_sns_topic" "gitlab_to_sns_webhook" {
  name                    = "${var.service_name}-gitlab-to-sns-webhook"
  kms_master_key_id = aws_kms_key.dev_kickstarter.key_id
  tags = var.tags
}

resource "aws_sqs_queue" "dev_kickstarter" {
  name                    = "${var.service_name}-dev-kickstarter"
  kms_master_key_id = aws_kms_key.dev_kickstarter.key_id
  kms_data_key_reuse_period_seconds = 300
  tags = var.tags
}

resource "aws_sns_topic_subscription" "dev_kickstarter_sns_subscription" {
  topic_arn = aws_sns_topic.gitlab_to_sns_webhook.arn
  protocol  = "sqs"
  endpoint  = aws_sqs_queue.dev_kickstarter.arn
  filter_policy = jsonencode({
    "eventName" : ["user_create"],
    "isBot" : ["false"] 
  })
}

data "aws_iam_policy_document" "dev_kickstarter_policy_document" {
  statement {
    effect = "Allow"

     actions = [
      "kms:Decrypt",
    ]

    resources = [
      aws_kms_key.dev_kickstarter.arn,
    ]
  }

  statement {
    effect = "Allow"

     actions = [
      "sqs:ReceiveMessage",
      "sqs:DeleteMessage",
      "sqs:GetQueueAttributes",
    ]

    resources = [
      aws_sqs_queue.dev_kickstarter.arn,
    ]
  }

  statement {
    effect = "Allow"
    
    actions = [
      "ses:SendEmail",
      "ses:SendRawEmail",
    ]

    resources = [
      "*"
    ]
  }
}