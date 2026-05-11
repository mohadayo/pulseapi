import logging
import os
from datetime import datetime, timezone

import httpx
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel, HttpUrl

logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s [%(levelname)s] %(message)s",
)
logger = logging.getLogger(__name__)

app = FastAPI(title="PulseAPI Monitor")

targets: dict[str, dict] = {}


class TargetCreate(BaseModel):
    name: str
    url: HttpUrl
    interval: int = 30


class TargetResponse(BaseModel):
    name: str
    url: str
    interval: int
    last_status: str | None = None
    last_checked: str | None = None


@app.get("/health")
def health():
    logger.info("Health check requested")
    return {
        "service": "monitor",
        "status": "healthy",
        "timestamp": datetime.now(timezone.utc).isoformat(),
    }


@app.get("/targets", response_model=list[TargetResponse])
def list_targets():
    logger.info("Listing all targets")
    return [
        TargetResponse(
            name=name,
            url=str(data["url"]),
            interval=data["interval"],
            last_status=data.get("last_status"),
            last_checked=data.get("last_checked"),
        )
        for name, data in targets.items()
    ]


@app.post("/targets", response_model=TargetResponse, status_code=201)
def add_target(target: TargetCreate):
    if target.name in targets:
        logger.warning("Target '%s' already exists", target.name)
        raise HTTPException(status_code=409, detail=f"Target '{target.name}' already exists")
    targets[target.name] = {
        "url": str(target.url),
        "interval": target.interval,
        "last_status": None,
        "last_checked": None,
    }
    logger.info("Added target '%s' -> %s", target.name, target.url)
    return TargetResponse(
        name=target.name,
        url=str(target.url),
        interval=target.interval,
    )


@app.delete("/targets/{name}", status_code=204)
def remove_target(name: str):
    if name not in targets:
        logger.warning("Target '%s' not found for deletion", name)
        raise HTTPException(status_code=404, detail=f"Target '{name}' not found")
    del targets[name]
    logger.info("Removed target '%s'", name)


@app.post("/targets/{name}/check")
def check_target(name: str):
    if name not in targets:
        logger.warning("Target '%s' not found for check", name)
        raise HTTPException(status_code=404, detail=f"Target '{name}' not found")

    url = targets[name]["url"]
    logger.info("Checking target '%s' at %s", name, url)

    try:
        with httpx.Client(timeout=5.0) as client:
            resp = client.get(f"{url}/health")
        status = "healthy" if resp.status_code == 200 else "unhealthy"
    except httpx.RequestError as e:
        logger.error("Failed to check '%s': %s", name, str(e))
        status = "unreachable"

    now = datetime.now(timezone.utc).isoformat()
    targets[name]["last_status"] = status
    targets[name]["last_checked"] = now
    logger.info("Target '%s' status: %s", name, status)

    return {
        "name": name,
        "status": status,
        "checked_at": now,
    }


if __name__ == "__main__":
    import uvicorn

    port = int(os.getenv("MONITOR_PORT", "8081"))
    logger.info("Starting monitor on port %d", port)
    uvicorn.run(app, host="0.0.0.0", port=port)
