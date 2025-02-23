# Web Traffic Simulator: A Real-time Event Processing System

## Overview

This project demonstrates the power of Go's concurrency model by simulating a high-throughput web traffic processing system, similar to what you might find in platforms like Reddit or Twitter. It showcases how to handle thousands of events per second while maintaining real-time processing capabilities.

## Why This Matters

The system demonstrates several critical concepts in modern software engineering:

1. **Concurrent Processing**: Shows how to handle multiple operations simultaneously without blocking
2. **Real-time Data Processing**: Demonstrates processing of live data streams
3. **Database Integration**: Shows proper handling of database operations in a concurrent environment
4. **Metrics and Monitoring**: Implements real-time visualization of system performance
5. **Production-like Architecture**: Mirrors real-world systems used by social media platforms

## System Architecture

The system consists of four main components running concurrently:

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│  Event          │     │   Database      │     │    Event        │
│  Generator      │────▶│   Writer        │────▶│    Processor    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
         │                      │                        │
         │                      │                        │
         ▼                      ▼                        ▼
┌───────────────────────────────────────────────────────────┐
│                    Metrics Visualizer                      │
└───────────────────────────────────────────────────────────┘
```

## Key Features

1. **Zero-Blocking Design**: All components operate independently without blocking each other
2. **Real-time Metrics**: Live visualization of system performance
3. **Database Integration**: PostgreSQL integration with proper concurrent access
4. **Graceful Shutdown**: Clean shutdown mechanism for all components
5. **Production-Ready Patterns**: Uses channels, mutexes, and goroutines in a production-like manner

## Technologies Used

- **Go**: For concurrent processing and system implementation
- **PostgreSQL**: For persistent storage
- **ANSI Colors**: For beautiful terminal visualization
- **Goroutines**: For concurrent execution
- **Channels**: For inter-process communication 