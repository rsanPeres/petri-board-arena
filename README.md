# 🧩 Petri Board Arena

**Petri Board Arena** is a **personal project** focused on **distributed systems, Apache Kafka, and event-driven architectures**, implemented in **Go** using **CQRS**, **Transactional Outbox**, **PostgreSQL**, and **Redis**.

This project was intentionally designed as a **hands-on architectural laboratory** to explore real-world backend challenges such as:

- Eventual consistency
- Asynchronous processing
- Reliable event delivery
- Read/write model separation
- Failure handling with retries and Dead Letter Queues (DLQ)

> ⚠️ **This is not a commercial product.**  
> This repository exists as a **learning-focused architecture project** aimed at demonstrating engineering decisions and trade-offs.

---

## 🎯 Project Goals

- Demonstrate **Kafka-based event pipelines**
- Apply **CQRS** beyond theoretical examples
- Implement the **Transactional Outbox pattern**
- Explore **consumer retry strategies and DLQs**
- Showcase clean and idiomatic **Go backend architecture**
- Serve as a **professional portfolio project**

---

## 🏗️ System Architecture (High Level)

The system is composed of the following main components:

1. **API (Go)**
   - Handles commands (write operations)
   - Persists domain data in PostgreSQL
   - Writes domain events to the Outbox table

2. **PostgreSQL (Write Model)**
   - Stores normalized domain data
   - Stores outbox events transactionally

3. **Outbox Worker**
   - Polls the outbox table
   - Publishes events to Kafka
   - Guarantees reliable event delivery

4. **Kafka**
   - Acts as the event backbone
   - Decouples producers and consumers
   - Enables reprocessing and fault isolation

5. **Projection Worker**
   - Consumes Kafka events
   - Applies projections to Redis
   - Handles retries and DLQ

6. **Redis (Read Model)**
   - Stores materialized views
   - Optimized for fast queries
   - Eventually consistent

> 📄 **Detailed architecture diagrams are available in [`docs/architecture.md`](docs/architecture.md).**

---

## 🧠 Key Architectural Decisions

### CQRS (Command Query Responsibility Segregation)

- Write model optimized for consistency
- Read model optimized for performance
- Asynchronous propagation via events

### Transactional Outbox

- Solves the dual-write problem
- Ensures no lost events
- Industry-proven pattern

### Kafka-Based Event Processing

- Durable event log
- Consumer groups
- Retry and DLQ handling

---

## 🧰 Technology Stack

- **Go** (API and worker)
- **PostgreSQL 16**
- **Redis 7**
- **Apache Kafka**
- **Docker & Docker Compose**
- **golang-migrate**
- **Makefile**

---

## 🚀 Running the Project Locally

### 1️⃣ Clone the repository

```bash
git clone git@github.com:petri-board-arena/petri-board-arena.git
cd petri-board-arena
```

2️⃣ Create the .env file
```bash
cp .env.example .env
```

3️⃣ Start infrastructure
```bash
make up
```
4️⃣ Apply database migrations
```bash
make update
```
5️⃣ Run the API
```bash
make dev
```
6️⃣ Run the worker
```bash
make worker
```
