package actions

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	mongoOptions "go.mongodb.org/mongo-driver/mongo/options"

	"github.com/JianLoong/robogo/internal/common"
	"github.com/JianLoong/robogo/internal/constants"
	"github.com/JianLoong/robogo/internal/types"
)

// mongodbAction handles MongoDB operations
func mongodbAction(args []any, options map[string]any, vars *common.Variables) types.ActionResult {
	// Validate arguments
	if len(args) < 3 {
		return types.MissingArgsError("mongodb", 3, len(args))
	}

	// Check for unresolved variables
	if errorResult := validateArgsResolved("mongodb", args); errorResult != nil {
		return *errorResult
	}

	operation := fmt.Sprintf("%v", args[0])
	connectionURL := fmt.Sprintf("%v", args[1])
	collection := fmt.Sprintf("%v", args[2])

	// Get timeout option
	timeout := 30 * time.Second
	if timeoutStr, ok := options["timeout"].(string); ok {
		if parsedTimeout, err := time.ParseDuration(timeoutStr); err == nil {
			timeout = parsedTimeout
		}
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Connect to MongoDB
	clientOptions := mongoOptions.Client().ApplyURI(connectionURL)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_CONNECTION_FAILED").
			WithTemplate("Failed to connect to MongoDB: %s").
			WithContext("connection_url", connectionURL).
			WithContext("error", err.Error()).
			WithSuggestion("Check if MongoDB is running and accessible").
			WithSuggestion("Verify connection string format").
			WithSuggestion("Check network connectivity").
			Build(err.Error())
	}
	defer client.Disconnect(ctx)

	// Test connection
	if err := client.Ping(ctx, nil); err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_PING_FAILED").
			WithTemplate("Failed to ping MongoDB: %s").
			WithContext("connection_url", connectionURL).
			WithContext("error", err.Error()).
			WithSuggestion("Check MongoDB server status").
			WithSuggestion("Verify authentication credentials").
			Build(err.Error())
	}

	// Execute operation
	switch operation {
	case "find":
		return executeMongoFind(ctx, client, collection, options)
	case "insert":
		return executeMongoInsert(ctx, client, collection, options)
	case "update":
		return executeMongoUpdate(ctx, client, collection, options)
	case "delete":
		return executeMongoDelete(ctx, client, collection, options)
	case "aggregate":
		return executeMongoAggregate(ctx, client, collection, options)
	case "count":
		return executeMongoCount(ctx, client, collection, options)
	default:
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "UNKNOWN_MONGODB_OPERATION").
			WithTemplate("Unknown MongoDB operation: %s").
			WithContext("operation", operation).
			WithContext("supported_operations", []string{"find", "insert", "update", "delete", "aggregate", "count"}).
			WithSuggestion("Use one of the supported operations: find, insert, update, delete, aggregate, count").
			Build(operation)
	}
}

// executeMongoFind handles find operations
func executeMongoFind(ctx context.Context, client *mongo.Client, collectionName string, options map[string]any) types.ActionResult {
	// Get database name from collection (format: "database.collection")
	dbName, collName := parseCollectionName(collectionName)
	collection := client.Database(dbName).Collection(collName)

	// Parse filter
	filter := bson.M{}
	if filterData, ok := options["filter"]; ok {
		filter = convertToBSON(filterData)
	}

	// Parse find options
	findOptions := mongoOptions.Find()

	if projection, ok := options["projection"]; ok {
		findOptions.SetProjection(convertToBSON(projection))
	}

	if limit, ok := options["limit"].(int); ok {
		findOptions.SetLimit(int64(limit))
	}

	if skip, ok := options["skip"].(int); ok {
		findOptions.SetSkip(int64(skip))
	}

	if sort, ok := options["sort"]; ok {
		findOptions.SetSort(convertToBSON(sort))
	}

	// Execute find
	cursor, err := collection.Find(ctx, filter, findOptions)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_FIND_FAILED").
			WithTemplate("MongoDB find operation failed: %s").
			WithContext("collection", collectionName).
			WithContext("filter", filter).
			WithContext("error", err.Error()).
			WithSuggestion("Check filter syntax and field names").
			WithSuggestion("Verify collection exists").
			Build(err.Error())
	}
	defer cursor.Close(ctx)

	// Decode results
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_DECODE_FAILED").
			WithTemplate("Failed to decode MongoDB results: %s").
			WithContext("error", err.Error()).
			Build(err.Error())
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"documents":  convertFromBSON(results),
			"count":      len(results),
			"collection": collectionName,
			"filter":     convertFromBSON(filter),
		},
	}
}

// executeMongoInsert handles insert operations
func executeMongoInsert(ctx context.Context, client *mongo.Client, collectionName string, options map[string]any) types.ActionResult {
	dbName, collName := parseCollectionName(collectionName)
	collection := client.Database(dbName).Collection(collName)

	// Get document(s) to insert
	if document, ok := options["document"]; ok {
		// Single document insert
		doc := convertToBSON(document)
		result, err := collection.InsertOne(ctx, doc)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_INSERT_FAILED").
				WithTemplate("MongoDB insert operation failed: %s").
				WithContext("collection", collectionName).
				WithContext("document", doc).
				WithContext("error", err.Error()).
				WithSuggestion("Check document format and required fields").
				WithSuggestion("Verify collection permissions").
				Build(err.Error())
		}

		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"inserted_id": convertFromBSON(result.InsertedID),
				"collection":  collectionName,
				"operation":   "insert_one",
			},
		}
	}

	if documents, ok := options["documents"].([]any); ok {
		// Multiple documents insert
		var docs []interface{}
		for _, doc := range documents {
			docs = append(docs, convertToBSON(doc))
		}

		result, err := collection.InsertMany(ctx, docs)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_INSERT_MANY_FAILED").
				WithTemplate("MongoDB insert many operation failed: %s").
				WithContext("collection", collectionName).
				WithContext("document_count", len(docs)).
				WithContext("error", err.Error()).
				Build(err.Error())
		}

		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"inserted_ids":   result.InsertedIDs,
				"inserted_count": len(result.InsertedIDs),
				"collection":     collectionName,
				"operation":      "insert_many",
			},
		}
	}

	return types.NewErrorBuilder(types.ErrorCategoryValidation, "MONGODB_MISSING_DOCUMENT").
		WithTemplate("MongoDB insert requires 'document' or 'documents' option").
		WithSuggestion("Add 'document' option for single insert").
		WithSuggestion("Add 'documents' option for multiple insert").
		Build("missing document data")
}

// executeMongoUpdate handles update operations
func executeMongoUpdate(ctx context.Context, client *mongo.Client, collectionName string, options map[string]any) types.ActionResult {
	dbName, collName := parseCollectionName(collectionName)
	collection := client.Database(dbName).Collection(collName)

	// Parse filter and update
	filter := bson.M{}
	if filterData, ok := options["filter"]; ok {
		filter = convertToBSON(filterData)
	}

	update, ok := options["update"]
	if !ok {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "MONGODB_MISSING_UPDATE").
			WithTemplate("MongoDB update requires 'update' option").
			WithSuggestion("Add 'update' option with update operations").
			Build("missing update data")
	}

	updateDoc := convertToBSON(update)

	// Check if it's update many or update one
	updateMany := false
	if many, ok := options["many"].(bool); ok {
		updateMany = many
	}

	if updateMany {
		result, err := collection.UpdateMany(ctx, filter, updateDoc)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_UPDATE_MANY_FAILED").
				WithTemplate("MongoDB update many operation failed: %s").
				WithContext("collection", collectionName).
				WithContext("filter", filter).
				WithContext("update", updateDoc).
				WithContext("error", err.Error()).
				Build(err.Error())
		}

		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"matched_count":  result.MatchedCount,
				"modified_count": result.ModifiedCount,
				"collection":     collectionName,
				"operation":      "update_many",
			},
		}
	} else {
		result, err := collection.UpdateOne(ctx, filter, updateDoc)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_UPDATE_FAILED").
				WithTemplate("MongoDB update operation failed: %s").
				WithContext("collection", collectionName).
				WithContext("filter", filter).
				WithContext("update", updateDoc).
				WithContext("error", err.Error()).
				Build(err.Error())
		}

		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"matched_count":  result.MatchedCount,
				"modified_count": result.ModifiedCount,
				"upserted_id":    result.UpsertedID,
				"collection":     collectionName,
				"operation":      "update_one",
			},
		}
	}
}

// executeMongoDelete handles delete operations
func executeMongoDelete(ctx context.Context, client *mongo.Client, collectionName string, options map[string]any) types.ActionResult {
	dbName, collName := parseCollectionName(collectionName)
	collection := client.Database(dbName).Collection(collName)

	// Parse filter
	filter := bson.M{}
	if filterData, ok := options["filter"]; ok {
		filter = convertToBSON(filterData)
	}

	// Check if it's delete many or delete one
	deleteMany := false
	if many, ok := options["many"].(bool); ok {
		deleteMany = many
	}

	if deleteMany {
		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_DELETE_MANY_FAILED").
				WithTemplate("MongoDB delete many operation failed: %s").
				WithContext("collection", collectionName).
				WithContext("filter", filter).
				WithContext("error", err.Error()).
				Build(err.Error())
		}

		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"deleted_count": result.DeletedCount,
				"collection":    collectionName,
				"operation":     "delete_many",
			},
		}
	} else {
		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_DELETE_FAILED").
				WithTemplate("MongoDB delete operation failed: %s").
				WithContext("collection", collectionName).
				WithContext("filter", filter).
				WithContext("error", err.Error()).
				Build(err.Error())
		}

		return types.ActionResult{
			Status: constants.ActionStatusPassed,
			Data: map[string]any{
				"deleted_count": result.DeletedCount,
				"collection":    collectionName,
				"operation":     "delete_one",
			},
		}
	}
}

// executeMongoAggregate handles aggregation operations
func executeMongoAggregate(ctx context.Context, client *mongo.Client, collectionName string, options map[string]any) types.ActionResult {
	dbName, collName := parseCollectionName(collectionName)
	collection := client.Database(dbName).Collection(collName)

	// Parse pipeline
	pipeline, ok := options["pipeline"].([]any)
	if !ok {
		return types.NewErrorBuilder(types.ErrorCategoryValidation, "MONGODB_MISSING_PIPELINE").
			WithTemplate("MongoDB aggregate requires 'pipeline' option").
			WithSuggestion("Add 'pipeline' option with aggregation stages").
			Build("missing pipeline data")
	}

	// Convert pipeline to BSON
	var bsonPipeline []bson.M
	for _, stage := range pipeline {
		bsonPipeline = append(bsonPipeline, convertToBSON(stage))
	}

	// Execute aggregation
	cursor, err := collection.Aggregate(ctx, bsonPipeline)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_AGGREGATE_FAILED").
			WithTemplate("MongoDB aggregation failed: %s").
			WithContext("collection", collectionName).
			WithContext("pipeline", bsonPipeline).
			WithContext("error", err.Error()).
			WithSuggestion("Check aggregation pipeline syntax").
			WithSuggestion("Verify field names and operators").
			Build(err.Error())
	}
	defer cursor.Close(ctx)

	// Decode results
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_AGGREGATE_DECODE_FAILED").
			WithTemplate("Failed to decode aggregation results: %s").
			WithContext("error", err.Error()).
			Build(err.Error())
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"results":    convertFromBSON(results),
			"count":      len(results),
			"collection": collectionName,
			"pipeline":   convertFromBSON(bsonPipeline),
		},
	}
}

// executeMongoCount handles count operations
func executeMongoCount(ctx context.Context, client *mongo.Client, collectionName string, options map[string]any) types.ActionResult {
	dbName, collName := parseCollectionName(collectionName)
	collection := client.Database(dbName).Collection(collName)

	// Parse filter
	filter := bson.M{}
	if filterData, ok := options["filter"]; ok {
		filter = convertToBSON(filterData)
	}

	// Execute count
	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return types.NewErrorBuilder(types.ErrorCategoryDatabase, "MONGODB_COUNT_FAILED").
			WithTemplate("MongoDB count operation failed: %s").
			WithContext("collection", collectionName).
			WithContext("filter", filter).
			WithContext("error", err.Error()).
			Build(err.Error())
	}

	return types.ActionResult{
		Status: constants.ActionStatusPassed,
		Data: map[string]any{
			"count":      count,
			"collection": collectionName,
			"filter":     convertFromBSON(filter),
		},
	}
}

// Helper functions

// parseCollectionName parses "database.collection" format
func parseCollectionName(fullName string) (string, string) {
	// If no dot, assume it's just collection name and use default database
	if len(fullName) == 0 {
		return "test", "test"
	}

	// Split by last dot to handle database names with dots
	for i := len(fullName) - 1; i >= 0; i-- {
		if fullName[i] == '.' {
			return fullName[:i], fullName[i+1:]
		}
	}

	// No dot found, use as collection name with default database
	return "test", fullName
}

// convertToBSON converts interface{} to bson.M recursively
func convertToBSON(data any) bson.M {
	switch v := data.(type) {
	case map[string]any:
		result := bson.M{}
		for key, value := range v {
			result[key] = convertToBSONValue(value)
		}
		return result
	case map[any]any:
		result := bson.M{}
		for key, value := range v {
			keyStr := fmt.Sprintf("%v", key)
			result[keyStr] = convertToBSONValue(value)
		}
		return result
	case bson.M:
		return v
	default:
		// If it's not a map, return empty bson.M (this should not happen in normal usage)
		return bson.M{}
	}
}

// convertToBSONValue converts individual values for BSON
func convertToBSONValue(value any) any {
	switch v := value.(type) {
	case map[string]any:
		result := bson.M{}
		for key, val := range v {
			result[key] = convertToBSONValue(val)
		}
		return result
	case map[any]any:
		result := bson.M{}
		for key, val := range v {
			keyStr := fmt.Sprintf("%v", key)
			result[keyStr] = convertToBSONValue(val)
		}
		return result
	case []any:
		var result []any
		for _, val := range v {
			result = append(result, convertToBSONValue(val))
		}
		return result
	default:
		return value
	}
}

// convertFromBSON converts MongoDB result types to JSON-compatible types
func convertFromBSON(data any) any {
	switch v := data.(type) {
	case bson.M:
		result := make(map[string]any)
		for key, value := range v {
			result[key] = convertFromBSON(value)
		}
		return result
	case []bson.M:
		var result []any
		for _, doc := range v {
			result = append(result, convertFromBSON(doc))
		}
		return result
	case []any:
		var result []any
		for _, item := range v {
			result = append(result, convertFromBSON(item))
		}
		return result
	case primitive.ObjectID:
		return fmt.Sprintf("ObjectID(\"%s\")", v.Hex())
	default:
		return v
	}
}
