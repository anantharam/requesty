# requesty
Testing LLM routing with requesty

This repo contains a sample backend for a LLM routing service using requesty and store the data in influxdb.
It has a swagger ui for testing the api.

This repo also contains a sample frontend for testing the api.
The frontend is a simple html page that allows you to send messages to the api and see the response using streaming or non-streaming.

This clearly shows the difference between streaming and non-streaming responses and how to use requesty to use the streaming responses :

## Streaming vs Non-Streaming Responses

When interacting with OpenAI's API, you can choose between streaming and non-streaming responses. Here's a comparison of their advantages:

### Streaming Advantages

1. **Improved User Experience**
   - Real-time response generation visible to users
   - Immediate feedback that the system is working
   - Users can start reading while the response is still being generated
   - Particularly valuable for long responses

2. **Reduced Perceived Latency**
   - First tokens appear quickly (typically within 100-200ms)
   - Better perceived responsiveness
   - Progressive content display
   - No need to wait for complete response generation

3. **Better Monitoring & Debugging**
   - Chunk-by-chunk monitoring capability
   - Detailed logging of streaming events
   - Easier to track performance metrics
   - Better visibility into the generation process

4. **Graceful Handling of Long Responses**
   - Prevents timeout issues with large responses
   - Progressive rendering of content
   - Better memory management
   - Reduced risk of connection issues

5. **Enhanced Interactivity**
   - Real-time UI updates
   - Dynamic content rendering
   - More engaging user experience
   - Immediate visual feedback

6. **Resource Efficiency**
   - Incremental processing of data
   - Better memory utilization
   - Optimized for mobile devices
   - Reduced server load for large responses

7. **Early Termination Option**
   - Users can stop generation if needed
   - Saves computation resources
   - Reduces wasted time
   - More user control

### When to Use Non-Streaming

Non-streaming responses might be preferred in these scenarios:

- Simple applications without real-time update requirements
- Batch processing scenarios
- When ensuring response completeness is critical
- When you need to process the entire response as a single unit
- For simpler applications where real-time updates aren't necessary
- When you need to perform validation or modification on the entire response before showing it to the user
- In scenarios where you want to ensure the complete response meets certain criteria before displaying anything

### Implementation

This repository supports both streaming and non-streaming modes through the `stream` parameter in requests.
The `stream` parameter in the `AnthropicRequest` struct makes it easy to toggle between the two modes:

```json
type AnthropicRequest struct {
    Stream bool `json:"stream,omitempty" example:"false"`
    // ... other fields
}
```

Choose the appropriate mode based on your specific use case and requirements.

