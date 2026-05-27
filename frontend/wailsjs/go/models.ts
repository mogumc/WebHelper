export namespace api {
	
	export class CookieInfo {
	    name: string;
	    value: string;
	    path: string;
	    domain: string;
	    expires: string;
	    maxAge: number;
	    secure: boolean;
	    httpOnly: boolean;
	    sameSite: string;
	
	    static createFrom(source: any = {}) {
	        return new CookieInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.value = source["value"];
	        this.path = source["path"];
	        this.domain = source["domain"];
	        this.expires = source["expires"];
	        this.maxAge = source["maxAge"];
	        this.secure = source["secure"];
	        this.httpOnly = source["httpOnly"];
	        this.sameSite = source["sameSite"];
	    }
	}
	export class HttpResponse {
	    statusCode: number;
	    status: string;
	    headers: Record<string, string>;
	    cookies: CookieInfo[];
	    body: string;
	    contentType: string;
	    contentLength: number;
	    time: string;
	    size: number;
	    error: string;
	
	    static createFrom(source: any = {}) {
	        return new HttpResponse(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.statusCode = source["statusCode"];
	        this.status = source["status"];
	        this.headers = source["headers"];
	        this.cookies = this.convertValues(source["cookies"], CookieInfo);
	        this.body = source["body"];
	        this.contentType = source["contentType"];
	        this.contentLength = source["contentLength"];
	        this.time = source["time"];
	        this.size = source["size"];
	        this.error = source["error"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class JsExecResult {
	    output: string;
	    result: string;
	    error: string;
	    duration: string;
	    success: boolean;
	
	    static createFrom(source: any = {}) {
	        return new JsExecResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.output = source["output"];
	        this.result = source["result"];
	        this.error = source["error"];
	        this.duration = source["duration"];
	        this.success = source["success"];
	    }
	}
	export class RequestLog {
	    id: number;
	    method: string;
	    url: string;
	    host: string;
	    path: string;
	    headers: Record<string, string>;
	    cookies: Record<string, string>;
	    bodyType: string;
	    body: string;
	    statusCode: number;
	    status: string;
	    respHeaders: Record<string, string>;
	    respBody: string;
	    respSize: number;
	    time: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new RequestLog(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.method = source["method"];
	        this.url = source["url"];
	        this.host = source["host"];
	        this.path = source["path"];
	        this.headers = source["headers"];
	        this.cookies = source["cookies"];
	        this.bodyType = source["bodyType"];
	        this.body = source["body"];
	        this.statusCode = source["statusCode"];
	        this.status = source["status"];
	        this.respHeaders = source["respHeaders"];
	        this.respBody = source["respBody"];
	        this.respSize = source["respSize"];
	        this.time = source["time"];
	        this.createdAt = source["createdAt"];
	    }
	}

}

export namespace global {
	
	export class LanguageInfo {
	    language_name: string;
	    language_code: string;
	    textmap_path: string;
	    translation_progress: string;
	    translator: string;
	    last_updated: string;
	    version: string;
	
	    static createFrom(source: any = {}) {
	        return new LanguageInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language_name = source["language_name"];
	        this.language_code = source["language_code"];
	        this.textmap_path = source["textmap_path"];
	        this.translation_progress = source["translation_progress"];
	        this.translator = source["translator"];
	        this.last_updated = source["last_updated"];
	        this.version = source["version"];
	    }
	}
	export class LanguagePack {
	    language_name: string;
	    language_code: string;
	    textmap_path: string;
	    translation_progress: string;
	    translator: string;
	    last_updated: string;
	    version: string;
	    textmap: Record<string, string>;
	
	    static createFrom(source: any = {}) {
	        return new LanguagePack(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.language_name = source["language_name"];
	        this.language_code = source["language_code"];
	        this.textmap_path = source["textmap_path"];
	        this.translation_progress = source["translation_progress"];
	        this.translator = source["translator"];
	        this.last_updated = source["last_updated"];
	        this.version = source["version"];
	        this.textmap = source["textmap"];
	    }
	}

}

export namespace service {
	
	export class ProxyConfig {
	    url: string;
	
	    static createFrom(source: any = {}) {
	        return new ProxyConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	    }
	}
	export class TimeoutConfig {
	    timeout: number;
	
	    static createFrom(source: any = {}) {
	        return new TimeoutConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.timeout = source["timeout"];
	    }
	}

}

