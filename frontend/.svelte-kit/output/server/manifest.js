export const manifest = (() => {
function __memo(fn) {
	let value;
	return () => value ??= (value = fn());
}

return {
	appDir: "_app",
	appPath: "_app",
	assets: new Set(["ort-wasm-simd-threaded.jsep.wasm","ort-wasm-simd-threaded.wasm","silero_vad_v5.onnx","vad.worklet.bundle.min.js"]),
	mimeTypes: {".wasm":"application/wasm",".js":"text/javascript"},
	_: {
		client: {start:"_app/immutable/entry/start.NtWCtgvF.js",app:"_app/immutable/entry/app.D3gDVt3_.js",imports:["_app/immutable/entry/start.NtWCtgvF.js","_app/immutable/chunks/CwPTCi79.js","_app/immutable/chunks/CEjoAh5C.js","_app/immutable/chunks/CDGuArWT.js","_app/immutable/entry/app.D3gDVt3_.js","_app/immutable/chunks/D9Z9MdNV.js","_app/immutable/chunks/CDGuArWT.js","_app/immutable/chunks/DsnmJJEf.js","_app/immutable/chunks/CEjoAh5C.js","_app/immutable/chunks/DcZzMn2F.js","_app/immutable/chunks/Bmg_b7yX.js","_app/immutable/chunks/BvYyoKJp.js"],stylesheets:[],fonts:[],uses_env_dynamic_public:false},
		nodes: [
			__memo(() => import('./nodes/0.js')),
			__memo(() => import('./nodes/1.js')),
			__memo(() => import('./nodes/4.js')),
			__memo(() => import('./nodes/7.js')),
			__memo(() => import('./nodes/8.js')),
			__memo(() => import('./nodes/12.js')),
			__memo(() => import('./nodes/15.js')),
			__memo(() => import('./nodes/18.js')),
			__memo(() => import('./nodes/20.js'))
		],
		remotes: {
			
		},
		routes: [
			{
				id: "/dev-server/[directoryId]",
				pattern: /^\/dev-server\/([^/]+?)\/?$/,
				params: [{"name":"directoryId","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 2 },
				endpoint: null
			},
			{
				id: "/files/[directoryId]",
				pattern: /^\/files\/([^/]+?)\/?$/,
				params: [{"name":"directoryId","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 3 },
				endpoint: null
			},
			{
				id: "/git/[directoryId]",
				pattern: /^\/git\/([^/]+?)\/?$/,
				params: [{"name":"directoryId","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 4 },
				endpoint: null
			},
			{
				id: "/projects/[id]",
				pattern: /^\/projects\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 5 },
				endpoint: null
			},
			{
				id: "/swarm/[session]",
				pattern: /^\/swarm\/([^/]+?)\/?$/,
				params: [{"name":"session","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 6 },
				endpoint: null
			},
			{
				id: "/task-executions/[id]",
				pattern: /^\/task-executions\/([^/]+?)\/?$/,
				params: [{"name":"id","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 7 },
				endpoint: null
			},
			{
				id: "/terminal/[session]",
				pattern: /^\/terminal\/([^/]+?)\/?$/,
				params: [{"name":"session","optional":false,"rest":false,"chained":false}],
				page: { layouts: [0,], errors: [1,], leaf: 8 },
				endpoint: null
			}
		],
		prerendered_routes: new Set(["/","/agents","/directories","/files","/kanban","/login","/projects","/settings","/swarm","/swarm/live","/task-executions","/terminal"]),
		matchers: async () => {
			
			return {  };
		},
		server_assets: {}
	}
}
})();
