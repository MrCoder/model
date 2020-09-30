import {defs} from "./defs";
import {create, setPosition} from "./svg-create";
import {cursorInteraction} from "svg-editor-tools/lib/cursor-interaction";
import {intersectRect, shapes} from "./shapes";



interface Group {
	id: string;
	name: string;
	nodes: (Node | Group)[];
	ref?: SVGGElement;

	x?: number;
	y?: number;
	width?: number;
	height?: number;
}

interface Node {
	id: string;
	title: string;
	sub: string;
	description: string;

	shape: string;

	ref?: SVGGElement;
	expanded?: boolean;
	x: number;
	y: number;
	width: number;
	height: number;
	selected?: boolean;

	intersect: (p: Point) => Point
}

interface Edge {
	label: string;
	from: Node;
	to: Node;
	ref?: SVGGElement;
	count: number;
}

interface Point {
	x: number;
	y: number;
}

export class GraphData {
	id: string;
	name: string;
	nodesMap: Map<string, Node>;
	edges: Edge[];
	edgeCounts: Map<string, number>;
	groupsMap: Map<string, Group>;

	constructor(id?: string, name?: string) {
		this.id = id;
		this.name = name;

		this.edges = [];
		this.nodesMap = new Map;
		this.groupsMap = new Map;
		this.edgeCounts = new Map;
	}

	addNode(id: string, label: string, sub: string, description: string, shape: string) {
		if (this.nodesMap.has(id)) throw Error('duplicate node: ' + id)
		const n: Node = {
			id, title: label, sub, description, shape,
			x: 0, y: 0, width: nodeWidth, height: nodeHeight, intersect: null
		}
		this.nodesMap.set(n.id, n)
	}

	nodes() {
		return Array.from(this.nodesMap.values())
	}

	addEdge(fromNode: string, toNode: string, label: string) {
		const id = `${fromNode}->${toNode}`
		if (this.edgeCounts.has(id)) {
			this.edgeCounts.set(id, this.edgeCounts.get(id) + 1)
		} else {
			this.edgeCounts.set(id, 1)
		}
		const e = {
			from: this.nodesMap.get(fromNode),
			to: this.nodesMap.get(toNode),
			label,
			count: this.edgeCounts.get(id)
		}
		this.edges.push(e)
	}

	addGroup(id: string, name: string, nodesOrGroups: string[]) {
		if (this.groupsMap.has(id)) {
			console.error(`Group exists: ${id} ${name}`)
			return
		}
		const group: Group = {
			id, name, nodes: nodesOrGroups.map(k => {
				const n = this.nodesMap.get(k) || this.groupsMap.get(k)
				if (!n) console.error(`Node or group ${k} not found for group ${id} "${name}"`)
				return n
			}).filter(Boolean)
		}
		this.groupsMap.set(id, group)
	}

	setExpanded(node: Node, ex: boolean) {
		node.expanded = ex;
		this.rebuildNode(node)
		updatePanning()
	}

	private rebuildNode(node: Node) {
		const p = node.ref.parentElement;
		p.removeChild(node.ref)
		node.ref = buildNode(node, this)
		p.appendChild(node.ref)
		this.redrawEdges(node)
		this.redrawGroups(node)
	}

	setSelected(nodes: Node[]) {
		this.nodesMap.forEach(n => {
			n.selected = false
			n.ref.classList.remove('selected')
		})
		nodes.forEach(n => {
			n.selected = true
			n.ref.classList.add('selected')
		});
		this.updateEdgesSel()
		//console.log(nodes.map(n => `'${n.name}'`).join(', '))
	}

	private updateEdgesSel() {
		this.edges.forEach(e => {
			if (e.to.selected || e.from.selected) {
				e.ref.classList.add('selected')
			} else {
				e.ref.classList.remove('selected')
			}
		})
	}

	moveNode(n: Node, x: number, y: number) {
		n.x = x;
		n.y = y;
		setPosition(n.ref, x, y)
		this.redrawEdges(n);
		this.redrawGroups(n)
	}

	//redraw connected edges
	private redrawEdges(n: Node) {
		this.edges.forEach(e => {
			if (e.from == n) {
				const p = e.ref.parentElement;
				p.removeChild(e.ref)
				e.ref = buildEdge(e)
				p.append(e.ref)
			}
			if (e.to == n) {
				const p = e.ref.parentElement;
				p.removeChild(e.ref)
				e.ref = buildEdge(e)
				p.append(e.ref)
			}
		})
		this.updateEdgesSel()
	}

	private redrawGroups(node: Node) {
		this.groupsMap.forEach(group => {
			//if (group.nodes.indexOf(node) == -1) return
			const p = group.ref.parentElement
			p.removeChild(group.ref)
			buildGroup(group)
			p.append(group.ref)
		})
	}

	//call this from console: JSON.stringify(gdata.exportLayout())
	exportLayout() {
		return Array.from(this.nodesMap.values())
			.reduce<{ [key: string]: { x: number, y: number } }>(
				(o, n) => {
					o[n.id] = {x: n.x, y: n.y};
					return o
				},
				{}
			)
	}

	exportSVG() {
		//save svg html
		let svg: SVGSVGElement = document.querySelector('svg#graph')
		const elastic = svg.querySelector('rect.elastic')
		const p = elastic.parentElement
		p.removeChild(elastic)
		const zoom = getZoom()
		setZoom(1)
		let src = svg.outerHTML
		p.append(elastic)
		setZoom(zoom)
		return src.replace(/^<svg/, '<svg xmlns="http://www.w3.org/2000/svg"')
	}

	importLayout(layout: { [key: string]: any }) {
		Object.entries(layout).forEach(([k, v]) => {
			const n = this.nodesMap.get(k)
			if (!n) return
			n.x = v.x
			n.y = v.y
			n.expanded = v.ex
		})
	}
}

let svg: SVGSVGElement = document.querySelector('svg#graph')
if (!svg) {
	svg = document.createElementNS("http://www.w3.org/2000/svg", "svg");
	svg.setAttribute('id', 'graph')
	svg.addEventListener('click', e => clickListener(e))
	addCursorInteraction(svg)
}
svg.setAttribute('width', '100%')
svg.setAttribute('height', '100%')

let clickListener: (e: MouseEvent) => void
let dragging = false;
let selectListener: (n: Node) => void


export const buildGraph = (data: GraphData, onNodeSelect: (n: Node) => void) => {
	// empty svg
	svg.innerHTML = defs
	document.body.append(svg) // make sure svg element is connected, we will measure texts sizes
	// @ts-ignore
	svg.__data = data

	selectListener = onNodeSelect

	//use event delegation
	clickListener = e => {
		if (dragging) {
			return;
		}
		// the expand button was clicked
		let el = (e.target as any).closest('.node > .expand');
		if (el) {
			const n: Node = el.parentElement.__data
			data.setExpanded(n, !n.expanded)
			return
		}
	}

	//toplevel groups
	const measureG = create.element('g', {}) as SVGGElement
	setPosition(measureG, 0, -100)
	svg.append(measureG)
	const zoomG = create.element('g', {}, 'zoom') as SVGGElement
	const nodesG = create.element('g', {}, 'nodes') as SVGGElement
	const edgesG = create.element('g', {}, 'edges') as SVGGElement
	const groupsG = create.element('g', {}, 'groups') as SVGGElement
	zoomG.append(groupsG, edgesG, nodesG)


	data.nodesMap.forEach((n) => {
		buildNode(n, data)
		nodesG.append(n.ref)
	})

	data.edges.forEach(e => {
		buildEdge(e)
		edgesG.append(e.ref)
	})

	data.groupsMap.forEach((group) => {
		buildGroup(group)
		groupsG.append(group.ref)
	})

	svg.append(zoomG)

	const elasticEl = create.rect(300, 300, 50, 50, 0, 'elastic')
	svg.append(elasticEl)

	return {
		svg,
		setZoom,
	}
}

function buildEdge(edge: Edge) {
	const n1 = edge.from, n2 = edge.to;
	let p0: Point, pn: Point, p1: Point, p2: Point, p3: Point, p4: Point;

	const overlap = (x1: number, w1: number, x2: number, w2: number) => !(x1 + w1 < x2 || x1 > x2 + w2)

	p0 = n1.intersect(n2)
	pn = n2.intersect(n1)

	const g = create.element('g', {}, 'edge') as SVGGElement

	// label
	const
		cx = (p0.x + pn.x) / 2,
		fontSize = 14,
		cy = (p0.y + pn.y) / 2 - 10 + ((edge.count - 1) * 50);
	let {txt, dy, maxW} = create.textArea(edge.label, 200, fontSize, false, cx, cy, 'middle')
	maxW += fontSize
	applyStyle(txt, styles.edgeText)
	const bbox = {x: cx - maxW / 2, y: cy, width: maxW, height: dy - 7}
	const bg = create.rect(bbox.width, bbox.height, bbox.x, bbox.y)
	applyStyle(bg, styles.edgeRect)
	g.append(bg, txt)

	// the path
	bbox.x += bbox.width / 2
	bbox.y += bbox.height / 2
	p1 = intersectRect(bbox, p0)
	p2 = intersectRect(bbox, pn)

	// const path = [p0, p1, p2, p3, p4, pn].filter(Boolean).map((p, i) => {
	// 	return (i == 0 ? 'M' : 'L') + `${p.x},${p.y}`
	// }).join(' ')
	const path = `M${p0.x},${p0.y} L${p1.x},${p1.y} M${p2.x},${p2.y} L${pn.x},${pn.y}`

	const p = create.path(path, {'marker-end': 'url(#arrow)'}, 'edge')
	applyStyle(p, styles.edgePath)
	g.append(p)


	edge.ref = g
	return g
}

const nodeWidth = 240;
const nodeHeight = 200;

function buildNode(n: Node, data: GraphData) {
	// @ts-ignore
	window.gdata = data

	const w = nodeWidth;//Math.max(60, textWidth(n.id), ...n.fields.map(f => textWidth(f.name))) + 70
	const h = nodeHeight;
	n.width = w;

	const g = create.element('g', {}, 'node') as SVGGElement
	n.selected && g.classList.add('selected')
	setPosition(g, n.x, n.y)

	const shapeFn = shapes[n.shape] || shapes.rect
	const shape: SVGElement = shapeFn(g, n);

	shape.classList.add('nodeBorder')
	applyStyle(shape, styles.nodeBorder)

	const tg = create.element('g') as SVGGElement
	let cy = Number(g.getAttribute('label-offset-y')) || 0
	{
		const fontSize = styles.nodeText1["font-size"]
		const bold = styles.nodeText1["font-weight"] == 'bold'
		const {txt, dy} = create.textArea(n.title, nodeWidth - 40, fontSize, bold, 0, cy, 'middle')
		applyStyle(txt, styles.nodeText)
		applyStyle(txt, styles.nodeText1)
		tg.append(txt)
		cy += dy
	}
	{
		const txt = create.text(`[${n.sub}]`, 0, cy, 'middle')
		applyStyle(txt, styles.nodeText)
		applyStyle(txt, styles.nodeText2)
		tg.append(txt)
		cy += 10
	}
	{
		cy += 10
		const {txt, dy} = create.textArea(n.description, nodeWidth - 40, 14, false, 0, cy, 'middle')
		applyStyle(txt, styles.nodeText)
		applyStyle(txt, styles.nodeText3)
		tg.append(txt)
		cy += dy
	}

	setPosition(tg, 0, -cy / 2)
	g.append(tg)

	// @ts-ignore
	g.__data = n;
	n.ref = g;

	return g
}


function buildGroup(group: Group) {
	if (group.nodes.length == 0) {
		return
	}
	const g = create.element('g', {}, 'group') as SVGGElement

	let p0: Point = {x: 1e100, y: 1e100}, p1: Point = {x: 0, y: 0}
	group.nodes.forEach(n => {
		const b = {x: n.x - n.width / 2, y: n.y - n.height / 2, width: n.width, height: n.height}
		p0.x = Math.min(p0.x, b.x)
		p0.y = Math.min(p0.y, b.y)
		p1.x = Math.max(p1.x, b.x + b.width)
		p1.y = Math.max(p1.y, b.y + b.height)
	})
	const pad = 25
	const w = Math.max(p1.x - p0.x, 200)
	const h = p1.y - p0.y + pad
	const bb = {
		x: p0.x - pad,
		y: p0.y - pad,
		width: w + pad * 2,
		height: h + pad * 2,
	}
	const r = create.rect(bb.width, bb.height, bb.x, bb.y)
	group.x = bb.x + bb.width/2
	group.y = bb.y + bb.height/2
	group.width = bb.width
	group.height = bb.height
	applyStyle(r, styles.groupRect)

	const txt = create.text(group.name, p0.x, p1.y + 30)
	applyStyle(txt, styles.groupText)

	g.append(r, txt)
	group.ref = g
}

function addCursorInteraction(svg: SVGSVGElement) {

	function getData(el: SVGElement) {
		// @ts-ignore
		return el.__data
	}

	const gd = () => (getData(svg) as GraphData)

	cursorInteraction({
		svg: svg,
		nodeFromEvent(e: MouseEvent): Node {
			e.preventDefault()
			let el = (e.target as SVGElement).closest('g.nodes g.node') as SVGElement
			return el && getData(el)
		},
		setSelection(nodes: Node[]) {
			gd().setSelected(nodes)
			selectListener(nodes[0])
		},
		setDragging(d: boolean) {
			dragging = d
		},
		isSelected(node: Node): boolean {
			return node.selected
		},
		getSelection(): Node[] {
			return gd().nodes().filter(n => n.selected)
		},
		getZoom: getZoom,
		moveNode(n: Node, x: number, y: number) {
			gd().moveNode(n, x, y)
		},
		boxSelection(box: DOMRect, add) {
			gd().setSelected(gd().nodes().filter(n => {
				return (add && n.selected) || svg.checkIntersection(n.ref.firstChild as SVGElement, box)
			}))
			selectListener(gd().nodes().find(n => n.selected))
		},
		updatePanning: updatePanning,
	})
}

function getZoom() {
	const el = svg.querySelector('g.zoom') as SVGGElement
	if (el.transform.baseVal.numberOfItems == 0) return 1
	return el.transform.baseVal.getItem(0).matrix.a
}

const svgPadding = 20

function setZoom(zoom: number) {
	const el = svg.querySelector('g.zoom') as SVGGElement
	el.setAttribute('transform', `scale(${zoom})`)
	// also set panning size
	updatePanning()
}

function updatePanning() {
	const el = svg.querySelector('g.zoom') as SVGGElement
	const bb = el.getBBox()
	const zoom = getZoom()
	const w = Math.max(svg.parentElement.clientWidth / zoom, bb.x + bb.width + svgPadding)
	const h = Math.max(svg.parentElement.clientHeight / zoom, bb.y + bb.height + svgPadding)
	svg.setAttribute('width', String(w * zoom))
	svg.setAttribute('height', String(h * zoom))
}

export const getZoomAuto = () => {
	const el = svg.querySelector('g.zoom') as SVGGElement
	const bb = el.getBBox()
	const zoom = Math.min(
		(svg.parentElement.clientWidth - 20) / (bb.width + bb.x + svgPadding),
		(svg.parentElement.clientHeight - 20) / (bb.height + bb.y + svgPadding),
	)
	return Math.max(Math.min(zoom, 1), .2)
}

const styles = {
	//node styles
	nodeBorder: {
		fill: "rgba(255, 255, 255, 0.86)",
		stroke: "#aaa",
		"stroke-width": "2px"
	},
	nodeText: {
		'font-family': 'Arial, sans-serif',
		fill: "#777",
		stroke: "none"
	},
	nodeText1: {
		'font-size': 16,
		'font-weight': 'bold',
	},
	nodeText2: {
		'font-size': 12,
	},
	nodeText3: {
		'font-size': 16,
	},

	//edge styles
	edgePath: {
		stroke: "#aaa",
		"stroke-width": 2,
	},
	edgeText: {
		stroke: "none",
		fill: "#777"
	},
	edgeRect: {
		fill: "none",
		stroke: "none",
	},

	//group styles
	groupRect: {
		// fill: "#ffc80026",
		fill: "rgba(255, 200, 0, 0.15)",
		stroke: "#666",
		"stroke-dasharray": 4,
	},
	groupText: {
		fill: "#666",
		"font-size": 18,
		cursor: "default"
	}
}

const applyStyle = (el: SVGElement, style: { [key: string]: string | number }) => {
	Object.keys(style).forEach(name => {
		if (name == 'font-size') {
			if (typeof (style[name]) != 'number') {
				console.error(`All font-sizes in styles have to be numbers representing px! Found:`, style)
			}
			el.setAttribute(name, style[name] + 'px')
		} else {
			el.setAttribute(name, String(style[name]))
		}
	})
}