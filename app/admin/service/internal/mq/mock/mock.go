package mock

import (
	"context"
	"fmt"

	"github.com/hypercoze/kratos-admin/app/admin/service/internal/mq"
)

type MockMQ struct{}

func (m MockMQ) Publish(ctx context.Context, msg *mq.Message, opts ...mq.Option) error {
	fmt.Println("mock publish:", string(msg.Body))
	return nil
}
