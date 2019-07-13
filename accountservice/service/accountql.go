package service

import (
	"github.com/sirupsen/logrus"
	internalmodel "github.com/callistaenterprise/goblog/accountservice/model"
	"github.com/callistaenterprise/goblog/common/model"
	"github.com/graphql-go/graphql"
	"log"
	"strconv"
	"time"
)

var schema graphql.Schema
var schemaInitialized = false

// define custom GraphQL ObjectType `todoType` for our Golang struct `Todo`
// Note that
// - the fields in our todoType maps with the json tags for the fields in our struct
// - the field type matches the field type in our struct

var accounts []internalmodel.Account

// init seeds some test data.
func init() {
	accounts = make([]internalmodel.Account, 0)
	for a := 0; a < 10; a++ {
		accountID := strconv.Itoa(120 + a)
		quote := internalmodel.Quote{Text: "HEJ", Language: "sv"}
		account := internalmodel.Account{ID: accountID, Name: "Test Testsson " + strconv.Itoa(a), ServedBy: "localhost", Quote: quote}
		account.AccountEvents = make([]model.AccountEvent, 0)
		account.AccountEvents = append(account.AccountEvents, model.AccountEvent{ID: strconv.Itoa(1), AccountID: accountID, EventName: "CREATED", Created: time.Now().Format("2006-01-02T15:04:05")})
		account.AccountEvents = append(account.AccountEvents, model.AccountEvent{ID: strconv.Itoa(2), AccountID: accountID, EventName: "UPDATED", Created: time.Now().Format("2006-01-02T15:04:05")})
		account.ImageData = model.AccountImage{ID: accountID, URL: "http://fake.path/image.png", ServedBy: "localhost"}
		accounts = append(accounts, account)
	}
}

func initQL(resolvers GraphQLResolvers) {
	if schemaInitialized {
		return
	}
	// ----------- Start declare types ------------------

	// quoteType
	var quoteType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Quote",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"quote": &graphql.Field{
				Type: graphql.String,
			},
			"ipAddress": &graphql.Field{
				Type: graphql.String,
			},
			"language": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// accountEventType
	var accountEventType = graphql.NewObject(graphql.ObjectConfig{
		Name: "AccountEvent",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"eventName": &graphql.Field{
				Type: graphql.String,
			},
			"created": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// accountImageType
	var accountImageType = graphql.NewObject(graphql.ObjectConfig{
		Name: "AccountImage",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"url": &graphql.Field{
				Type: graphql.String,
			},
		},
	})

	// accountType, includes Resolver functions for inner quotes and events.
	var accountType = graphql.NewObject(graphql.ObjectConfig{
		Name: "Account",
		Fields: graphql.Fields{
			"id": &graphql.Field{
				Type: graphql.String,
			},
			"name": &graphql.Field{
				Type: graphql.String,
			},
			"servedBy": &graphql.Field{
				Type: graphql.String,
			},
			"quote": &graphql.Field{
				Type: quoteType,
				Args: graphql.FieldConfigArgument{
					"language": &graphql.ArgumentConfig{
						Type:        graphql.String,
						Description: "Two letter ISO language code such as en or sv",
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					logrus.Infof("ENTER - resolve function for quote with params %v", p.Args)
					account := p.Source.(internalmodel.Account)
					if account.Quote.Language == p.Args["language"] {
						return account.Quote, nil
					}
					return account.Quote, nil
				},
			},
			"events": &graphql.Field{
				Type: graphql.NewList(accountEventType),
				Args: graphql.FieldConfigArgument{
					"eventName": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					logrus.Infof("ENTER - resolve function for events with params %v", p.Args)
					account := p.Source.(internalmodel.Account)

					if len(p.Args) == 0 {
						return account.AccountEvents, nil
					}

					response := make([]model.AccountEvent, 0)
					for _, item := range account.AccountEvents {
						if item.EventName == p.Args["eventName"] {
							response = append(response, item)
						}
					}
					return response, nil
				},
			},
			"imageData": &graphql.Field{
				Type: accountImageType,
				Resolve: func(p graphql.ResolveParams) (interface{}, error) {
					logrus.Infof("ENTER - resolve function for imageData with params %v", p.Args)
					account := p.Source.(internalmodel.Account)
					return account.ImageData, nil
				},
			},
		},
	})

	// Schema
	fields := graphql.Fields{
		"Account": &graphql.Field{
			Type: graphql.Type(accountType),
			Args: graphql.FieldConfigArgument{
				"id": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
				"name": &graphql.ArgumentConfig{
					Type: graphql.String,
				},
			},
			Resolve: resolvers.AccountResolverFunc,
		},
		//"AllAccounts": &graphql.Field{
		//    Type: graphql.NewList(accountType),
		//    Args: graphql.FieldConfigArgument{
		//        "id": &graphql.ArgumentConfig{
		//            Type: graphql.String,
		//        },
		//        "name": &graphql.ArgumentConfig{
		//            Type: graphql.String,
		//        },
		//    },
		//    Resolve: resolvers.AllAccountsResolverFunc,
		//},
	}

	rootQuery := graphql.ObjectConfig{Name: "RootQuery", Fields: fields}
	schemaConfig := graphql.SchemaConfig{Query: graphql.NewObject(rootQuery)}
	var err error
	schema, err = graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}
	logrus.Infoln("Successfully initialized GraphQL")
	schemaInitialized = true
}

// GetAccountQL is a HTTP handler function that decodes the supplied Base64-encoded GraphQL query, executes the supplied
// GraphQL query and returns the result as JSON.
//func GetAccountQLBase64(resp http.ResponseWriter, req *http.Request) {
//    logrus.Infoln("ENTER - GetAccountQLBase64")
//    query := req.URL.Query()["query"]
//    logrus.Infof("Read query %v\n", query)
//    decodedString, err := base64.URLEncoding.DecodeString(query[0])
//    logrus.Infof("Decoded query: %v\n", string(decodedString))
//
//    if err != nil {
//        writeJSONResponse(resp, 400, []byte(err.Error()))
//        return
//    }
//
//    params := graphql.Params{Schema: schema, RequestString: string(decodedString)}
//    r := graphql.Do(params)
//    if len(r.Errors) > 0 {
//        logrus.Errorf("failed to execute graphql operation, errors: %+v", r.Errors)
//        log.Fatalf("failed to execute graphql operation, errors: %+v", r.Errors)
//    }
//    rJSON, _ := json.Marshal(r)
//    writeJSONResponse(resp, 200, rJSON)
//}
