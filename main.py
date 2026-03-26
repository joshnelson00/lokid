from fastapi import FastAPI, Request
from fastapi.templating import Jinja2Templates
from fastapi.responses import HTMLResponse
from fastapi.staticfiles import StaticFiles
from chaos import get_all_nodes, get_all_pods
from kubernetes import client, config

app = FastAPI()
templates = Jinja2Templates(directory="templates")
app.mount("/static", StaticFiles(directory="static"), name="static")

config.load_kube_config()
api = client.CoreV1Api()

@app.get("/")
async def root(request: Request):
    nodes = get_all_nodes(api)
    pods = get_all_pods(api)
    return templates.TemplateResponse(
        request=request,
        name="index.html",
        context={"name": "Josh", "nodes": nodes, "pods": pods}
    ) 