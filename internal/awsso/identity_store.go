package awsso

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/identitystore/types"
)

type IdentityStore struct {
	client *identitystore.Client

	IdentityStoreID string  `json:"identity_store_id"`
	Groups          []Group `json:"groups"`
	Users           []User  `json:"users"`
}

type Group struct {
	GroupID     string   `json:"group_id"`
	Description string   `json:"description"`
	DisplayName string   `json:"display_name"`
	Members     []string `json:"members"`
}

type User struct {
	UserID      string `json:"user_id"`
	UserName    string `json:"user_name"`
	FamilyName  string `json:"family_name"`
	GivenName   string `json:"given_name"`
	DisplayName string `json:"display_name"`
}

func NewIdentityStore(ctx context.Context, identityStoreID string) (*IdentityStore, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return &IdentityStore{
		client:          identitystore.NewFromConfig(cfg),
		IdentityStoreID: identityStoreID,
	}, nil
}

func (s *IdentityStore) GetUsers(ctx context.Context) error {
	s.Users = make([]User, 0, 8192)

	params := &identitystore.ListUsersInput{
		IdentityStoreId: aws.String(s.IdentityStoreID),
		MaxResults:      aws.Int32(50),
	}

	paginator := identitystore.NewListUsersPaginator(s.client, params, func(o *identitystore.ListUsersPaginatorOptions) {
		o.Limit = 50
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, v := range output.Users {
			s.Users = append(s.Users, User{
				UserID:      aws.ToString(v.UserId),
				UserName:    aws.ToString(v.UserName),
				FamilyName:  aws.ToString(v.Name.FamilyName),
				GivenName:   aws.ToString(v.Name.GivenName),
				DisplayName: aws.ToString(v.DisplayName),
			})
		}
	}

	sort.SliceStable(s.Users, func(i, j int) bool { return s.Users[i].UserName < s.Users[j].UserName })

	return nil
}

func (s *IdentityStore) GetGroups(ctx context.Context) error {
	s.Groups = make([]Group, 0, 1024)

	params := &identitystore.ListGroupsInput{
		IdentityStoreId: aws.String(s.IdentityStoreID),
		MaxResults:      aws.Int32(50),
	}

	paginator := identitystore.NewListGroupsPaginator(s.client, params, func(o *identitystore.ListGroupsPaginatorOptions) {
		o.Limit = 50
	})

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return err
		}

		for _, v := range output.Groups {
			s.Groups = append(s.Groups, Group{
				GroupID:     aws.ToString(v.GroupId),
				Description: aws.ToString(v.Description),
				DisplayName: aws.ToString(v.DisplayName),
			})
		}
	}

	sort.SliceStable(s.Groups, func(i, j int) bool { return s.Groups[i].DisplayName < s.Groups[j].DisplayName })

	return nil
}

func (s *IdentityStore) getMembers(ctx context.Context, groupID string) ([]string, error) {
	members := make([]string, 0, 256)

	params := &identitystore.ListGroupMembershipsInput{
		IdentityStoreId: aws.String(s.IdentityStoreID),
		GroupId:         aws.String(groupID),
		MaxResults:      aws.Int32(50),
	}

	paginator := identitystore.NewListGroupMembershipsPaginator(
		s.client, params, func(o *identitystore.ListGroupMembershipsPaginatorOptions) {
			o.Limit = 50
		},
	)

	for paginator.HasMorePages() {
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range output.GroupMemberships {
			switch val := v.MemberId.(type) {
			case *types.MemberIdMemberUserId:
				members = append(members, val.Value)
			case *types.UnknownUnionMember:
				return nil, fmt.Errorf("unknown tag: %s", val.Tag)
			default:
				return nil, fmt.Errorf("member id is nil or unknown type")
			}
		}
	}

	return members, nil
}

func (s *IdentityStore) GetMembers(ctx context.Context) error {
	userMap := make(map[string]string, len(s.Users))

	for _, v := range s.Users {
		if _, ok := userMap[v.UserID]; !ok {
			userMap[v.UserID] = v.UserName
		} else {
			return fmt.Errorf("user id already exsists: %s", v.UserID)
		}
	}

	for n, v := range s.Groups {
		members, err := s.getMembers(ctx, v.GroupID)
		if err != nil {
			return err
		}

		s.Groups[n].Members = make([]string, 0, 256)
		for _, m := range members {
			if val, ok := userMap[m]; ok {
				s.Groups[n].Members = append(s.Groups[n].Members, val)
			} else {
				return fmt.Errorf("not found user id: %s", m)
			}
		}

		sort.SliceStable(s.Groups[n].Members, func(i, j int) bool { return s.Groups[n].Members[i] < s.Groups[n].Members[j] })
	}

	return nil
}
