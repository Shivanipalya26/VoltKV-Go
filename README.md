
# GoCache

A simplified, educational  implementation of an in-memory key-value store, built in Go. It is designed to mimic some of the core features of [Redis](https://redis.io/) , focusing on understanding how key-value stores work and how data is serialized and communicated using the Redis Serialization Protocol (RESP). 

---

## 📦 Features

- Custom RESP protocol parser
- In-memory key-value store
- Key expiration support 
- Handles multiple clients over TCP
- Basic Redis-like commands

---

## 🚀 Getting Started

### 1. **Clone the Repository**
```bash
git clone https://github.com/Shivanipalya26/go-redis.git
cd go-redis
```

### 2. **Install Go**

Make sure Go is installed. If not, download it from [golang.org](https://golang.org/dl/).

```bash
go version
```

### 3. **Run the Server**
```bash
go run main.go
```

By default, the server starts on port `6379`.

---

## 🛠️ Project Structure

```bash
├── main.go         # Entry point: starts the server
├── server/         # Connection handling 
├── internals/           
    ├── resp/       # RESP parsing & RESP encoder utilities
    ├── store/      # In-memory key-value store & expiration logic
├── cmd/            # Command executor logic
```

---

## 🧪 Example Usage

Use `redis-cli` in Docker or a Redis client to test the server:

```bash
docker run -it --rm redis redis-cli -h <your-host-ip> -p 6379
```

Then type:
```bash
PING
→ PONG

SET mykey "Hello"
→ OK

GET mykey
→ "Hello"

DEL mykey
→ (integer) 1

EXISTS mykey
→ (integer) 0
```

---

## 🔧 Supported Commands

| Command        | Description                        |
|----------------|------------------------------------|
| `PING [msg]`   | Responds with `PONG` or msg        |
| `SET <k> <v>`  | Sets a key-value pair              |
| `GET <k>`      | Gets the value for the key         |
| `DEL <k1>`     | Deletes key                        |
| `MSET k1 v1..` | Sets multiple keys                 |
| `MGET k1 k2..` | Gets multiple keys                 |
| `HSET <k> <f> <v>`   | Sets a field in a hash     |
| `HGET <k> <field>`   | Gets the value of a field in a hash         |
| `HGETALL <k>`        | Returns all fields and values of a hash     |
| `EXPIRE <k> <sec>` | Set TTL for a key             |
| `EXISTS <k>`      | Checks if the key exists      |
| `LPUSH <k> <v1>..`   | Pushes one or more values to the left       |
| `RPUSH <k> <v1>..`   | Pushes one or more values to the right      |
| `LPOP <k>`           | Removes and returns the first element       |
| `RPOP <k>`           | Removes and returns the last element        |

---

## 🧑‍💻 Contribution

Feel free to fork the repo and submit pull requests!  
This project is made for learning and experimentation.