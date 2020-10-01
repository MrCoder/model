import React, {FC, useState} from "react";
import {getZoomAuto, GraphData} from "./graph-view/graph";
import {Graph} from "./graph-view/graph-react";
import {BrowserRouter as Router, Route} from 'react-router-dom'
import {useHistory} from "react-router";
import {listViews, parseView, ViewsList} from "./parseModel";


export const Root: FC<{model: any, layout: any}> = ({model, layout}) => <Router>
	<Route path="/" component={() => <ModelPane key={getCrtID()} model={model} layouts={layout}/>}/>
</Router>

const getCrtID = () => {
	const p = new URLSearchParams(document.location.search)
	return p.get('id') || ''
}

const DomainSelect: FC<{ views: ViewsList; crtID: string}> = ({views, crtID}) => {
	const history = useHistory();
	return <select
		onChange={e => history.push('?id=' + encodeURIComponent(e.target.value))} value={crtID}>
		<option disabled value="" hidden>...</option>
		{views.map(m => <option key={m.key} value={m.key}>{m.section + ' ' + m.title}</option>)}
	</select>
}

const ModelPane: FC<{model: any, layouts: any}> = ({model, layouts}) => {
	const crtID = getCrtID()
	const [zoom, setZoom] = useState(1)
	const [saving, setSaving] = useState(false)

	const [graph] = useState(parseView(model, layouts, crtID))
	if (!graph) {
		return <div style={{padding:30}}><DomainSelect views={listViews(model)} crtID=""/></div>
	}

	function saveLayout() {
		setSaving(true)

		fetch('data/save?id=' + encodeURIComponent(crtID), {
			method: 'post',
			body: JSON.stringify({
				layout: graph.exportLayout(),
				svg: graph.exportSVG()
			})
		}).then(ret => {
			if (ret.status != 202) {
				alert('Error saving')
			}
			setSaving(false)
		})
	}
	return <>
		<div className="toolbar">
			<div>
				View: <DomainSelect views={listViews(model)} crtID={crtID}/>
				{' '}
			</div>
			<div>
				<button onClick={() => setZoom(zoom - .05)}>Zoom -</button>
				{' '}
				<button onClick={() => setZoom(zoom + .05)}>Zoom +</button>
				{' '}
				<button onClick={() => setZoom(getZoomAuto())}>Fit</button>
				{' '}
				<button onClick={() => setZoom(1)}>Zoom 100%</button>
				{' '}
				<button className="action" disabled={saving} onClick={() => saveLayout()}>Save View</button>
			</div>
		</div>
		<Graph key={crtID}
			   data={graph}
			   zoom={zoom}
			   // print metadata in console
			   onSelect={id => id && console.log(removeEmptyProps(graph.metadata.elements.find((m: any) => m.id == id)))}
			   onInit={() => setZoom(getZoomAuto())}/>
	</>
}

function removeEmptyProps(o: any) {
	return JSON.parse(JSON.stringify(o))
}