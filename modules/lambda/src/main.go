package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/google/uuid"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// dynamodb session
type db struct {
	svc *dynamodb.DynamoDB
}

func client() *db {
	// create dynamodb client
	sess := session.Must(session.NewSession())
	return &db{
		svc: dynamodb.New(sess),
	}
}

// extracts the JSON and writes it to DynamoDB
func post(body string) (map[string]interface{}, error) {
	db := client()
	// unmarshal the request body
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		log.Fatal(err)
	}
	// PutItem in table
	err := db.put(data)
	return data, err
}

func (s *db) put(data map[string]interface{}) error {
	vv, err := dynamodbattribute.ConvertToMap(data)
	if err != nil {
		log.Println("Failed to convert to dynamodb attributes")
		return err
	}
	// generate unique id
	id := uuid.NewString()
	vv["id"] = &dynamodb.AttributeValue{S: &(id)}
	// write to dynamodb
	params := &dynamodb.PutItemInput{
		Item:      vv,
		TableName: aws.String(os.Getenv("TABLE_NAME")), // Required
	}
	_, err = s.svc.PutItem(params)
	return err
}

func handler(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// log body
	log.Println("Received body: ", request.Body)

	resource := request.Path[1:]

	// handle empty resource path
	if resource == "" {
		log.Println("No resource specified")
		return events.APIGatewayProxyResponse{
			Body:       "Error: no FHIR resource specified in path (e.g. /Patient)\n",
			StatusCode: 400,
		}, nil
	}

	log.Println("Resource: ", resource)
	// check if server accepts resource type
	accepted := AcceptedResource(resource)
	if accepted {
		// write to dynamodb
		item, err := post(request.Body)
		if err != nil {
			log.Println("Error calling post() %e", err.Error())

			return events.APIGatewayProxyResponse{
				Body:       "Error",
				StatusCode: 500,
			}, nil
		}

		// log and return result
		log.Println("Wrote item: ", item)
		return events.APIGatewayProxyResponse{
			Body:       "Success \n",
			StatusCode: 200,
		}, nil
	}
	// not accepted resource type
	log.Println("Invalid resource type")
	return events.APIGatewayProxyResponse{
		Body:       "Error: invalid or unsupported FHIR resource\n",
		StatusCode: 405,
	}, nil

}

func main() {
	lambda.Start(handler)
}
