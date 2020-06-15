package grpc

import (
	"context"
	"encoding/hex"
	"fmt"
	"github.com/gloompi/tantora-back/app/dbConnection"
	"github.com/gloompi/tantora-back/app/proto/tantorapb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var connection = dbConnection.ReadConnection()

type Server struct{}

func (*Server) Friends(_ context.Context, req *tantorapb.FriendsRequest) (*tantorapb.FriendsResponse, error) {
	userId := req.GetUserId()

	if len(userId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Received an empty userId")
	}

	query := fmt.Sprintf(`
		select
			f.friend_id,
			u.user_name,
			u.first_name,
			u.last_name
		from friends f
			inner join users as u
			on f.friend_id = u.user_id
		where f.user_id = %v;
	`, userId)

	rows, err := connection.DB.Query(query)
	if err != nil {
		return nil, err
	}

	var friends []*tantorapb.Friend

	for rows.Next() {
		friend := &tantorapb.Friend{}

		err := rows.Scan(
			&friend.FriendId,
			&friend.UserName,
			&friend.FirstName,
			&friend.LastName,
		)

		if err != nil {
			return nil, err
		}

		friends = append(friends, friend)
	}

	res := &tantorapb.FriendsResponse{
		Friends: friends,
	}

	return res, nil
}

func (*Server) RecentMessages(_ context.Context, req *tantorapb.RecentMessagesRequest) (*tantorapb.RecentMessagesResponse, error) {
	userId := req.GetUserId()

	if len(userId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Received empty userId")
	}

	query := fmt.Sprintf(`
		select
			case
				when receiver_id = %v then sender_id
				else receiver_id
			end as receiver_id,
			created_date,
			u.user_name,
			u.first_name,
			u.last_name 
		from
			(select distinct on(receiver_id) receiver_id, sender_id, created_date
			from message
			where sender_id = %v or receiver_id = 14%v
			order by receiver_id, created_date desc) message
			inner join users as u on
			(case
				when receiver_id = %v then sender_id = u.user_id
				when sender_id = %v then receiver_id = user_id
			end)
		order by created_date desc;
	`, userId, userId, userId, userId, userId)

	rows, err := connection.DB.Query(query)
	if err != nil {
		return nil, err
	}

	var messages []*tantorapb.RecentMessage

	for rows.Next() {
		recentMes := &tantorapb.RecentMessage{}

		err := rows.Scan(
			&recentMes.UserId,
			&recentMes.UserName,
			&recentMes.FirstName,
			&recentMes.LastName,
			&recentMes.CreatedDate,
		)

		if err != nil {
			return nil, err
		}

		messages = append(messages, recentMes)
	}

	res := &tantorapb.RecentMessagesResponse{
		RecentMessages: messages,
	}

	return res, nil
}

func (*Server) Messages(_ context.Context, req *tantorapb.ChatRequest) (*tantorapb.ChatResponse, error) {
	userId := req.GetUserId()
	receiverId := req.GetReceiverId()
	limit := req.GetLimit()
	offset := req.GetOffset()

	if len(userId) == 0 || len(receiverId) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "Received an empty userId or receiverId")
	}

	if limit == 0{
		limit = 10
	}

	query := fmt.Sprintf(`
		select
			m.sender_id,
			m.receiver_id,
			m."content",
			m.created_date
		from message m
		where m.sender_id = %v and m.receiver_id = %v or m.sender_id = %v and m.receiver_id = %v
		order by m.created_date desc
		limit %v offset %v;
	`, userId, receiverId, receiverId, userId, limit, offset)

	rows, err := connection.DB.Query(query)
	if err != nil {
		return nil, err
	}

	var messages []*tantorapb.ChatMessage

	for rows.Next() {
		message := &tantorapb.ChatMessage{}

		err := rows.Scan(
			&message.SenderId,
			&message.ReceiverId,
			&message.Content,
			&message.CreatedDate,
		)

		if err != nil {
			return nil, err
		}

		decodedStr, _ := hex.DecodeString(message.Content)
		message.Content = string(decodedStr)
		messages = append(messages, message)
	}

	res := &tantorapb.ChatResponse{
		Messages: messages,
	}

	query = fmt.Sprintf(`
		select user_name, first_name, last_name
		from users
		where user_id = %v;
	`, receiverId)

	rows, err = connection.DB.Query(query)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		err := rows.Scan(
			&res.UserName,
			&res.FirstName,
			&res.LastName,
		)

		if err != nil {
			return nil, err
		}
	}

	return res, nil
}

func (*Server) SaveMessage(_ context.Context, req *tantorapb.SaveMessageRequest) (*tantorapb.SaveMessageResponse, error) {
	message := req.GetMessage()

	if message == nil {
		return nil, status.Errorf(codes.InvalidArgument, "Received an empty `message`")
	}

	query := fmt.Sprintf(`
		insert into message (
			sender_id,
			receiver_id,
			"content"
		) values ('%v', '%v', '%v');
	`, message.GetSenderId(), message.GetReceiverId(), hex.EncodeToString([]byte(message.GetContent())))

	stmt, err := connection.DB.Prepare(query)
	if err != nil {
		return nil, err
	}

	defer stmt.Close()

	_, err = stmt.Exec()

	res := &tantorapb.SaveMessageResponse{
		Status: 0,
	}

	if err == nil {
		res.Status = 1
	}

	return res, err
}
