# Garden-Runc

Two Revolutionary New Products: A runC backend for garden. And a garden front-end for runC.

Garden-Runc is a backend for Garden (which means it works in Cloud Foundry) and provides a super-small, production-quality runC manager.

# Technical Goals (Potentially)

 - Small and simple
 - No long-running daemon
 - Containers survive restarts
 - Docker plugin compatibility (maybe, someday)

# Why would I use this rather than docker?

Garden-runc is a small, non-opinionated wrapper around runC. This means the orchestrator of garden-runc has a lot
more flexibility than if it were orchestrating a full docker daemon. Also there's no long-running daemon and containers survive
upgrades of garden-runc. In other words, all the flexibility and simplicity of runC, all the convenience and power of a long-running daemon (but 
the daemon can restart without affecting running containers).
