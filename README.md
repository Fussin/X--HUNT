# SentinelX

**AI-Powered Elite Recon & Exploitation Proxy**

SentinelX is a next-generation, AI-powered, fully autonomous, enterprise-grade web proxy, vulnerability scanning, and exploitation framework.

## 🚀 Project Vision

Our vision is to build a system that can:

-   Intercept, modify, and analyze HTTP/HTTPS traffic in real-time.
-   Scan and exploit millions of domains/webapps per minute using distributed clusters.
-   Include AI Hacker Agents that autonomously map attack surfaces, generate payloads, and validate vulnerabilities.
-   Auto-generate PoCs with screenshots, request/response logs, and PDF reports.

## ✨ What's New in SentinelX v6.0

SentinelX v6.0 introduces a revolutionary multi-agent architecture with autonomous AI agents, intelligent decision-making, and advanced vulnerability intelligence.

### 🎯 Major Enhancements & New Capabilities
- **🤖 12+ AI Agents:** Autonomous penetration testing specialists.
- **🛠️ 150+ Security Tools:** Complete enterprise security arsenal.
- **🧠 Intelligent Decision Engine:** AI-powered strategy and optimization.
- **🎨 Modern Visual Engine:** Beautiful real-time dashboards.
- **⚡ Advanced Process Management:** Smart caching & resource optimization.
- **🔍 Vulnerability Intelligence:** CVE analysis & exploit generation.

### 🤖 Autonomous AI Agents (NEW!)
- **IntelligentDecisionEngine:** AI-powered tool selection and parameter optimization.
- **BugBountyWorkflowManager:** Specialized workflows for bug bounty hunting.
- **CTFWorkflowManager:** Automated CTF challenge solving with category-specific approaches.
- **CVEIntelligenceManager:** Real-time vulnerability intelligence and exploit analysis.
- **AIExploitGenerator:** Automated exploit development from CVE data.
- **VulnerabilityCorrelator:** Multi-stage attack chain discovery and optimization.
- **TechnologyDetector:** Advanced technology stack identification and analysis.
- **RateLimitDetector:** Intelligent rate limiting detection and timing adjustment.
- **FailureRecoverySystem:** Automatic error handling and alternative tool selection.
- **PerformanceMonitor:** Real-time system optimization and resource allocation.
- **ParameterOptimizer:** Context-aware parameter optimization for maximum effectiveness.
- **GracefulDegradation:** Fault-tolerant operation with partial tool failures.

### 🎨 Modern Visual Engine (ENHANCED!)
- **Reddish Hacker Theme:** Professional cybersecurity aesthetic with blood-red accents.
- **Real-time Progress Bars:** Beautiful animated progress indicators with ETA calculations.
- **Live Dashboards:** Multi-process monitoring with system metrics and health status.
- **Vulnerability Cards:** Color-coded severity indicators with detailed risk analysis.
- **Enhanced Logging:** Emoji-rich, color-coded output with structured formatting.
- **Terminal Animations:** Smooth animations and visual feedback for all operations.

### ⚡ Advanced Process Management (NEW!)
- **Smart Caching System:** Intelligent result caching with LRU eviction and TTL optimization.
- **Process Pools:** Auto-scaling thread pools with intelligent resource allocation.
- **Command Termination:** Real-time process control without server restart.
- **Resource Monitoring:** CPU, memory, and network usage optimization.
- **Error Recovery:** Automatic retry mechanisms with exponential backoff.
- **Performance Analytics:** Detailed metrics and optimization recommendations.

### 🔍 Vulnerability Intelligence System (NEW!)
- **CVE Real-time Monitoring:** Automated CVE feed analysis with severity filtering.
- **Exploitability Analysis:** AI-powered assessment of vulnerability exploitability.
- **Attack Chain Discovery:** Multi-stage attack path identification and optimization.
- **Threat Intelligence:** IOC correlation across multiple threat intelligence sources.
- **Zero-Day Research:** Automated vulnerability pattern recognition and analysis.
- **Exploit Database Integration:** Real-time exploit availability checking and correlation.

### 🛠️ Expanded Tool Arsenal (50+ NEW TOOLS!)
- **Network Security:** Rustscan, Masscan, AutoRecon, NetExec, Responder
- **Web Application:** Katana, HTTPx, Feroxbuster, Arjun, ParamSpider, X8, Jaeles, Dalfox
- **Cloud Security:** Prowler, Scout Suite, CloudMapper, Pacu, Trivy, Kube-Hunter, Kube-Bench
- **Binary Analysis:** Ghidra, Radare2, Pwntools, ROPgadget, One_gadget, Angr, Volatility3
- **API Testing:** GraphQL introspection, JWT manipulation, REST API fuzzing, WebSocket testing
- **CTF Specialized:** Advanced cryptography tools, steganography detection, forensics suite
- **OSINT & Reconnaissance:** Advanced subdomain enumeration, social media analysis, breach data correlation

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
