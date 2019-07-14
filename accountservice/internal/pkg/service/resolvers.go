package service

import (
	"fmt"
	"github.com/graphql-go/graphql"
	"github.com/sirupsen/logrus"
)

type GraphQLResolvers interface {
	AccountResolverFunc(p graphql.ResolveParams) (interface{}, error)
}

// LiveGraphQLResolvers implementations
type LiveGraphQLResolvers struct {
}

func (gqlres *LiveGraphQLResolvers) AccountResolverFunc(p graphql.ResolveParams) (interface{}, error) {
	account, err := fetchAccount(p.Context, p.Args["id"].(string))
	if err != nil {
		return nil, err
	}
	return account, nil
}

func (gqlres *LiveGraphQLResolvers) AllAccountsResolverFunc(p graphql.ResolveParams) (interface{}, error) {
	panic("implement me")
}

// TestGraphQLResolvers test implementations
type TestGraphQLResolvers struct {
}

func (gqlres *TestGraphQLResolvers) AccountResolverFunc(p graphql.ResolveParams) (interface{}, error) {
	logrus.Infof("ENTER - resolve function for Account with params %v", p.Args)
	id, _ := p.Args["id"].(string)
	for _, account := range accounts {
		if account.ID == id {
			return account, nil
		}
	}
	return nil, fmt.Errorf("No account found matching ID %v", id)
}

func (gqlres *TestGraphQLResolvers) AllAccountsResolverFunc(p graphql.ResolveParams) (interface{}, error) {
	logrus.Infof("ENTER - resolve function for AllAccounts with params %v", p.Args)
	return accounts, nil
}
