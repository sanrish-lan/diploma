class VDOMComponent:
    def __init__(self, tag: str, content: str = "", **attributes):
        self.tag = tag
        self.content = content
        self.attributes = attributes
        self.children = []

    def add_child(self, child):
        self.children.append(child)

    def render(self) -> str:
        attrs = " ".join(f'{k}="{v}"' for k, v in self.attributes.items())
        attr_str = f" {attrs}" if attrs else ""
        children_html = "".join(c.render() for c in self.children)
        return f"<{self.tag}{attr_str}>{self.content}{children_html}</{self.tag}>"


class VDOMPage:
    def __init__(self, title: str):
        self.title = title
        self.root = VDOMComponent("div", style="font-family: sans-serif; padding: 20px;")

    def add(self, component: VDOMComponent):
        self.root.add_child(component)

    def to_html(self) -> str:
        return f"""<!DOCTYPE html>
	<html>
	<head><title>{self.title}</title></head>
	<body style="background: #f4f6f8;">
    	{self.root.render()}
	</body>
	</html>"""
