package mq

import (
	"context"
	"sync"
	"time"

	"example.com/demo/pkg/log"

	ali_mns "github.com/aliyun/aliyun-mns-go-sdk"
)

type MNSConfig struct {
	Endpoint        string `json:"endpoint"       mapstructure:"endpoint"`
	AccessKeyID     string `json:"access_key_id"  mapstructure:"access_key_id"`
	AccessKeySecret string `json:"access_key_secret"  mapstructure:"access_key_secret"`
}

type AliyunMNS struct {
	client ali_mns.MNSClient
}

var consumeMNS sync.WaitGroup

func NewMNSConfig() *MNSConfig {
	return &MNSConfig{
		Endpoint:        "",
		AccessKeyID:     "",
		AccessKeySecret: "",
	}
}

func NewAliyunMNS(config *MNSConfig) MQ {

	client := ali_mns.NewAliMNSClient(config.Endpoint, config.AccessKeyID, config.AccessKeySecret)

	return &AliyunMNS{client: client}
}

func (mq *AliyunMNS) SendMessage(ctx context.Context, queueName string, messages ...*Message) error {
	queue := ali_mns.NewMNSQueue(queueName, mq.client)
	for _, msg := range messages {
		priority := msg.Priority
		if msg.Priority == 0 {
			priority = 8
		}

		reqMsg := ali_mns.MessageSendRequest{
			MessageBody:  msg.Content,
			DelaySeconds: int64(msg.DelaySeconds),
			Priority:     int64(priority),
		}
		ret, err := queue.SendMessage(reqMsg)
		if err != nil {
			log.Errorf("Send message failed! Error message is %s", err)
			return err
		}

		log.Infof("Send message success! MessageId is %s", ret.MessageId)

	}

	return nil
}

func (mq *AliyunMNS) ConsumeMessage(ctx context.Context, c Consumer, callback func(interface{}) bool) error {

	endChan := make(chan struct{})
	respChan := make(chan ali_mns.MessageReceiveResponse)
	errChan := make(chan error)
	queue := ali_mns.NewMNSQueue(c.QueueName, mq.client)

	// id := uuid.Must(uuid.NewV4(), nil).String()
	consumeMNS.Add(1)
	defer consumeMNS.Done()

	go func() {
		log.Infof("start to receive message from queue %s", c.QueueName)
	outerLoop:
		for {
			select {
			case <-ctx.Done():
				endChan <- struct{}{}
				break outerLoop
			default:
				queue.ReceiveMessage(respChan, errChan, 10)
				// log.Infof("Receive message from queue %s", c.QueueName)
			}
		}
	}()

	for {

		select {
		case <-endChan:
			log.Infof("end to receive message from queue %s", c.QueueName)
			return nil
		case resp := <-respChan:
			log.Infof("Receive message success! MessageBody is %s", resp.MessageBody)
			callback([]byte(resp.MessageBody))
			if err := queue.DeleteMessage(resp.ReceiptHandle); err != nil {
				log.Errorf("Delete message failed! Error message is %s", err)
			}
		case err := <-errChan:
			if ali_mns.ERR_MNS_MESSAGE_NOT_EXIST.IsEqual(err) {
				// log.Info("No message!")
			} else if ali_mns.ERR_MNS_QUEUE_NOT_EXIST.IsEqual(err) {
				log.Errorf("%s Queue is not exist!", c.QueueName)
				time.Sleep(5 * time.Second)
			} else {
				log.Errorf("Receive message failed! Error message is %s", err)
			}

		}
	}

	// return nil
}

func (mq *AliyunMNS) Close() error {
	consumeMNS.Wait()
	log.Info("close AliMNSMQ")
	return nil
}
