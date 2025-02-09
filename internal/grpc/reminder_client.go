package grpc

import (
	"context"

	"github.com/PharmaKart/gateway-svc/internal/proto"
	"google.golang.org/grpc"
)

type ReminderClient interface {
	ScheduleReminder(ctx context.Context, req *proto.ScheduleReminderRequest) (*proto.ScheduleReminderResponse, error)
	ListReminders(ctx context.Context, req *proto.ListRemindersRequest) (*proto.ListRemindersResponse, error)
	ListCustomerReminders(ctx context.Context, req *proto.ListCustomerRemindersRequest) (*proto.ListRemindersResponse, error)
	UpdateReminder(ctx context.Context, req *proto.UpdateReminderRequest) (*proto.UpdateReminderResponse, error)
	DeleteReminder(ctx context.Context, req *proto.DeleteReminderRequest) (*proto.DeleteReminderResponse, error)
	ToggleReminder(ctx context.Context, req *proto.ToggleReminderRequest) (*proto.ToggleReminderResponse, error)
	ListReminderLogs(ctx context.Context, req *proto.ListReminderLogsRequest) (*proto.ListReminderLogsResponse, error)
}

type reminderClient struct {
	client proto.ReminderServiceClient
}

func NewReminderServiceClient(conn *grpc.ClientConn) ReminderClient {
	return &reminderClient{
		client: proto.NewReminderServiceClient(conn),
	}
}

func (c *reminderClient) ScheduleReminder(ctx context.Context, req *proto.ScheduleReminderRequest) (*proto.ScheduleReminderResponse, error) {
	return c.client.ScheduleReminder(ctx, req)
}

func (c *reminderClient) ListReminders(ctx context.Context, req *proto.ListRemindersRequest) (*proto.ListRemindersResponse, error) {
	return c.client.ListReminders(ctx, req)
}

func (c *reminderClient) ListCustomerReminders(ctx context.Context, req *proto.ListCustomerRemindersRequest) (*proto.ListRemindersResponse, error) {
	return c.client.ListCustomerReminders(ctx, req)
}

func (c *reminderClient) UpdateReminder(ctx context.Context, req *proto.UpdateReminderRequest) (*proto.UpdateReminderResponse, error) {
	return c.client.UpdateReminder(ctx, req)
}

func (c *reminderClient) DeleteReminder(ctx context.Context, req *proto.DeleteReminderRequest) (*proto.DeleteReminderResponse, error) {
	return c.client.DeleteReminder(ctx, req)
}

func (c *reminderClient) ToggleReminder(ctx context.Context, req *proto.ToggleReminderRequest) (*proto.ToggleReminderResponse, error) {
	return c.client.ToggleReminder(ctx, req)
}

func (c *reminderClient) ListReminderLogs(ctx context.Context, req *proto.ListReminderLogsRequest) (*proto.ListReminderLogsResponse, error) {
	return c.client.ListReminderLogs(ctx, req)
}
