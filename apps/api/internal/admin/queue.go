package admin

import (
	"log"
	"net/http"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type QueueSize struct {
	Size int `json:"size"`
}

func NewQueueSizeHandler(client *sqs.Client, queueURL string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		output, err := client.GetQueueAttributes(r.Context(), &sqs.GetQueueAttributesInput{
			QueueUrl:       &queueURL,
			AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameApproximateNumberOfMessages},
		})
		if err != nil {
			log.Printf("failed to get queue attributes: %v", err)
			http.Error(w, "failed to load queue size", http.StatusInternalServerError)
			return
		}

		size := 0
		if raw, ok := output.Attributes[string(types.QueueAttributeNameApproximateNumberOfMessages)]; ok {
			if parsed, err := strconv.Atoi(raw); err == nil {
				size = parsed
			}
		}

		writeJSON(w, QueueSize{Size: size})
	}
}
