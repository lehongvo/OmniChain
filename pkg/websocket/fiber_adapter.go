package websocket

// Note: This file is kept for reference
// For Fiber framework, use nhooyr.io/websocket instead of gorilla/websocket
// because gorilla/websocket requires standard net/http which Fiber doesn't use

// Example with nhooyr.io/websocket:
//
// import "nhooyr.io/websocket"
//
// func HandleWebSocket(c *fiber.Ctx) error {
//     conn, err := websocket.Accept(c.Context(), c.Response(), c.Request(), nil)
//     if err != nil {
//         return err
//     }
//     // Use conn...
// }
