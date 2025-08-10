# SentinelX

**AI-Powered Elite Recon & Exploitation Proxy**

SentinelX is a next-generation, AI-powered, fully autonomous, enterprise-grade web proxy, vulnerability scanning, and exploitation framework.

## 🚀 Project Vision

Our vision is to build a system that can:

-   Intercept, modify, and analyze HTTP/HTTPS traffic in real-time.
-   Scan and exploit millions of domains/webapps per minute using distributed clusters.
-   Include AI Hacker Agents that autonomously map attack surfaces, generate payloads, and validate vulnerabilities.
-   Auto-generate PoCs with screenshots, request/response logs, and PDF reports.

## 📦 Monorepo Structure

This repository is a monorepo containing the source code for all SentinelX components.

-   `core-proxy/`: The core MITM engine written in Go.
-   `ai-hacker/`: AI agents and automation scripts written in Python.
-   `vuln-scanner/`: The vulnerability scanning engine.
-   `poc-engine/`: The Proof of Concept generation engine.
-   `zero-day-detector/`: Anomaly-based zero-day detection engine.
-   `distributed-mesh/`: Cluster orchestration for distributed scanning.
-   `web-ui/`: The main web dashboard and backend API.
-   `plugin-sdk/`: The SDK for developing third-party plugins.
-   `connectors/`: Integrations with platforms like HackerOne and Bugcrowd.
-   `compliance/`: Security and legal logging modules.
-   `docs/`: Project documentation.
-   `protos/`: Protobuf definitions for the gRPC message bus.

## Quick Start

> **Note:** This project is under active development. The following instructions are a target for the final setup.

1.  **Run the system:** `docker-compose up`
2.  **Access the UI:** [http://localhost:3000](http://localhost:3000)

---

*This project was bootstrapped with the help of an AI software engineer.*
