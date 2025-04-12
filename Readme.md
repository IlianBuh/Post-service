
# 📝 gRPC Post Service Documentation

## 🌐 Service Overview
**Post Service** is ф gRPC microservice for managing blog posts with file attachments. It provides efficient binary communication using Protocol Buffers.

## 📋 Method Specifications

### 📝 Post Management
| gRPC Method | Request | Response | Description |
|-------------|---------|----------|-------------|
| `Create` | `CreateRequest` | `CreateResponse` | Creates new blog post |
| `Update` | `UpdateRequest` | `UpdateResponse` | Updates existing post |
| `Delete` | `DeleteRequest` | `DeleteResponse` | Deletes specified post |

### 📁 File Operations
| gRPC Method | Request | Response | Description |
|-------------|---------|----------|-------------|
| `UploadFile` | `UploadFileRequest` | `UploadFileResponse` | Attaches file to post |
| `DownloadFile` | `DownloadFileRequest` | `DownloadFileResponse` | Retrieves attached file |


## 📊 Message Structures

### CreatePost Operation
```protobuf
message CreateRequest {
    int64 userId = 2;
    string header = 3;
    string content = 4;
    repeated string themes = 5;
}
message CreateResponse{}
```

### File Upload Operation
```protobuf
message UploadFileRequest {
    int64 postId = 1;
    string fileName = 2;
    bytes fileData = 3;
}
message UploadFileResponse{
    string fileUrl = 1;
}
```

