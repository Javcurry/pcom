package iamqp

import "time"

// Publish ...
func (s *Session) Publish(exchange, key string, msg Message, opt ...PubOpt) error {
	pub, err := s.publisher()
	if err != nil {
		return err
	}
	defer pub.Close()
	return pub.publish(exchange, key, msg, opt...)
}

// PublishWithConfirm publish messages using confirm mode. This function will blocked when
// message doesn't ack, and retry sending message FOREVER.
func (s *Session) PublishWithConfirm(exchange, key string, msg Message, opt ...PubOpt) error {
	pub, err := s.publisher()
	if err != nil {
		return err
	}
	defer pub.Close()
	for {
		err := pub.publish(exchange, key, msg, opt...)
		if err != nil {
			Log("iamqp: publish failed. retrying...")
			select {
			case <-s.done:
				return ErrShutDown
			case <-time.After(resendDelay):
			}
			continue
		}
		select {
		case confirm := <-pub.notifyConfirm:
			if confirm.Ack {
				Log("iamqp: publish confirmed!")
				return nil
			}
		case <-time.After(resendDelay):
		}
	}
}
