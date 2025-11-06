package core

import (
	"testing"

	. "github.com/onsi/gomega"

	"go.fd.io/govpp/api"
	interfaces "go.fd.io/govpp/binapi/interface"
)

func TestNotificationMessageAfterUnsubscribe(t *testing.T) {
	ctx := setupTest(t)
	defer ctx.teardownTest()

	notifMsg := &interfaces.SwInterfaceEvent{}
	msgID, err := ctx.conn.GetMessageID(notifMsg)
	Expect(err).ShouldNot(HaveOccurred())

	// Before subscribing, isNotificationMessage should return false
	isNotif := ctx.conn.isNotificationMessage(msgID)
	Expect(isNotif).Should(BeFalse())

	// Subscribe to the notification
	notifChan := make(chan api.Message, 10)
	sub, err := ctx.ch.SubscribeNotification(notifChan, notifMsg)
	Expect(err).ShouldNot(HaveOccurred())
	Expect(sub).ShouldNot(BeNil())

	// After subscribing, isNotificationMessage should return true
	isNotif = ctx.conn.isNotificationMessage(msgID)
	Expect(isNotif).Should(BeTrue())

	// Unsubscribe from the notification
	err = sub.Unsubscribe()
	Expect(err).ShouldNot(HaveOccurred())

	// After unsubscribing the last (and only) subscription, isNotificationMessage should return false
	isNotif = ctx.conn.isNotificationMessage(msgID)
	Expect(isNotif).Should(BeFalse(), "isNotificationMessage should return false after unsubscribing from the last subscription")
}
