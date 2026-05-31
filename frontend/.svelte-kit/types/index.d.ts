type DynamicRoutes = {
	"/dev-server/[directoryId]": { directoryId: string };
	"/files/[directoryId]": { directoryId: string };
	"/git/[directoryId]": { directoryId: string };
	"/projects/[id]": { id: string };
	"/swarm/[session]": { session: string };
	"/task-executions/[id]": { id: string };
	"/terminal/[session]": { session: string }
};

type Layouts = {
	"/": { directoryId?: string; id?: string; session?: string };
	"/agents": undefined;
	"/dev-server": { directoryId?: string };
	"/dev-server/[directoryId]": { directoryId: string };
	"/directories": undefined;
	"/files": { directoryId?: string };
	"/files/[directoryId]": { directoryId: string };
	"/git": { directoryId?: string };
	"/git/[directoryId]": { directoryId: string };
	"/kanban": undefined;
	"/login": undefined;
	"/projects": { id?: string };
	"/projects/[id]": { id: string };
	"/settings": undefined;
	"/swarm": { session?: string };
	"/swarm/live": undefined;
	"/swarm/[session]": { session: string };
	"/task-executions": { id?: string };
	"/task-executions/[id]": { id: string };
	"/terminal": { session?: string };
	"/terminal/[session]": { session: string }
};

export type RouteId = "/" | "/agents" | "/dev-server" | "/dev-server/[directoryId]" | "/directories" | "/files" | "/files/[directoryId]" | "/git" | "/git/[directoryId]" | "/kanban" | "/login" | "/projects" | "/projects/[id]" | "/settings" | "/swarm" | "/swarm/live" | "/swarm/[session]" | "/task-executions" | "/task-executions/[id]" | "/terminal" | "/terminal/[session]";

export type RouteParams<T extends RouteId> = T extends keyof DynamicRoutes ? DynamicRoutes[T] : Record<string, never>;

export type LayoutParams<T extends RouteId> = Layouts[T] | Record<string, never>;

export type Pathname = "/" | "/agents" | "/dev-server" | `/dev-server/${string}` & {} | "/directories" | "/files" | `/files/${string}` & {} | "/git" | `/git/${string}` & {} | "/kanban" | "/login" | "/projects" | `/projects/${string}` & {} | "/settings" | "/swarm" | "/swarm/live" | `/swarm/${string}` & {} | "/task-executions" | `/task-executions/${string}` & {} | "/terminal" | `/terminal/${string}` & {};

export type ResolvedPathname = `${"" | `/${string}`}${Pathname}`;

export type Asset = "/ort-wasm-simd-threaded.jsep.wasm" | "/ort-wasm-simd-threaded.wasm" | "/silero_vad_v5.onnx" | "/vad.worklet.bundle.min.js";