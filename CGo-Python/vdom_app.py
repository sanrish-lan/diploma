from vdom_models import VDOMComponent, VDOMPage

VISITORS_LOG = []

def process_request(req: dict) -> dict:
    """
    Принимает от Go словарь вида:
    {
        "method": "GET",
        "path": "/users",
        "headers": {"User-Agent": "...", ...}
    }
    Возвращает dict с кодом ответа и сгенерированным HTML.
    """
    method = req.get("method", "GET")
    path = req.get("path", "/")
    headers = req.get("headers", {})

    VISITORS_LOG.append((method, path))

    page = VDOMPage(title=f"VDOM 3.14 - {path}")

    card = VDOMComponent("div", style="background: white; border-radius: 8px; padding: 24px; box-shadow: 0 2px 8px rgba(0,0,0,0.1);")
    card.add_child(VDOMComponent("h1", "VDOM Server-Side Engine (Python 3.14)"))
    card.add_child(VDOMComponent("p", f"Запрос обработан через связку Go -> C-API -> Python Classes"))

    info_list = VDOMComponent("ul")
    info_list.add_child(VDOMComponent("li", f"HTTP Method: <b>{method}</b>"))
    info_list.add_child(VDOMComponent("li", f"Request Path: <b>{path}</b>"))
    info_list.add_child(VDOMComponent("li", f"Всего запросов в сессии Python: <b>{len(VISITORS_LOG)}</b>"))
    card.add_child(info_list)

    card.add_child(VDOMComponent("h3", "Заголовки запроса (переданы из Go в dict):"))
    headers_list = VDOMComponent("ul", style="font-size: 0.9em; color: #555;")
    for k, v in headers.items():
        headers_list.add_child(VDOMComponent("li", f"<b>{k}:</b> {v}"))
    card.add_child(headers_list)

    page.add(card)

    return {
        "status": 200,
        "content_type": "text/html; charset=utf-8",
        "body": page.to_html()
    }
