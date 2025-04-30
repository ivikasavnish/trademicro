# Python Process Manager (gRPC)

This is a Python-based process manager service that exposes a gRPC API for starting, stopping, resuming, and listing Python trading scripts on the destination machine.

## Features
- Start, stop, resume, and list Python trading processes
- Tracks running processes in memory
- Communicates via gRPC (port 50051)

## Usage

### 1. Install dependencies
```
pip install grpcio grpcio-tools
```

### 2. Generate gRPC code from proto
```
python -m grpc_tools.protoc -I.. --python_out=. --grpc_python_out=. ../process_manager.proto
```

### 3. Run the server
```
python server.py
```

### 4. Example gRPC client usage
See the proto file for method signatures. Use any gRPC client (Python, Go, etc.) to call `StartProcess`, `StopProcess`, `ResumeProcess`, and `ListProcesses`.

## Security
- Only exposes the gRPC API on the local network by default (port 50051)
- No authentication by default—add TLS or network firewall as needed

## Extending
- Add persistent process tracking or logging as needed
- Add authentication or authorization for production use
