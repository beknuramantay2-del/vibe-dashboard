export namespace main {

	export class FileChangeDTO {
	    path: string;
	    additions: number;
	    deletions: number;

	    static createFrom(source: any = {}) {
	        return new FileChangeDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
	    }
	}

	export class ResultDTO {
	    ok: boolean;
	    message: string;

	    static createFrom(source: any = {}) {
	        return new ResultDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.message = source["message"];
	    }
	}

	export class SessionDTO {
	    id: string;
	    agent: string;
	    project: string;
	    status: string;
	    cost: number;
	    inputTokens: number;
	    outputTokens: number;
	    cacheHitRate: number;
	    duration: string;
	    startTime: string;
	    pid: number;

	    static createFrom(source: any = {}) {
	        return new SessionDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.agent = source["agent"];
	        this.project = source["project"];
	        this.status = source["status"];
	        this.cost = source["cost"];
	        this.inputTokens = source["inputTokens"];
	        this.outputTokens = source["outputTokens"];
	        this.cacheHitRate = source["cacheHitRate"];
	        this.duration = source["duration"];
	        this.startTime = source["startTime"];
	        this.pid = source["pid"];
	    }
	}

	export class SnapshotDTO {
	    id: string;
	    sessionId: string;
	    message: string;
	    createdAt: string;

	    static createFrom(source: any = {}) {
	        return new SnapshotDTO(source);
	    }

	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.sessionId = source["sessionId"];
	        this.message = source["message"];
	        this.createdAt = source["createdAt"];
	    }
	}

}
