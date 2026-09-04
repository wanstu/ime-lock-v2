export namespace main {
	
	export class AppState {
	    running: boolean;
	    auto_fix: boolean;
	    auto_start: boolean;
	    silent_start: boolean;
	    capture_logs: boolean;
	    fix_count: number;
	    last_fix_at: string;
	    last_error: string;
	    shortcut: string;
	    config_path: string;
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.auto_fix = source["auto_fix"];
	        this.auto_start = source["auto_start"];
	        this.silent_start = source["silent_start"];
	        this.capture_logs = source["capture_logs"];
	        this.fix_count = source["fix_count"];
	        this.last_fix_at = source["last_fix_at"];
	        this.last_error = source["last_error"];
	        this.shortcut = source["shortcut"];
	        this.config_path = source["config_path"];
	    }
	}

}

