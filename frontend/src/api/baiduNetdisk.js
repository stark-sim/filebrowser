import Vue from "vue";
import store from "@/store";
import router from "@/router";
import { baseURL } from "@/utils/constants";
import { fetchURL, removePrefix } from "@/api/utils";
import i18n from "@/i18n";

/* util  */
async function fetchUtil(
  url,
  opts,
  setAt = true,
  atExpired = "baiduNetdisk.authExpired"
) {
  opts = opts || {};
  opts.headers = opts.headers || {};
  opts.body = opts.method === "GET" ? undefined : opts.body || {};

  let { headers, body, ...rest } = opts;
  if (setAt && body instanceof Object) {
    const at = getToken();
    if (!at) return;
    body = { ...body, access_token: at };
  }
  body = typeof body === "string" ? body : JSON.stringify(body);
  let res;

  try {
    res = await fetch(`${baseURL}${url}`, {
      headers: {
        "Content-Type": "application/json",
        ...headers,
      },
      body,
      ...rest,
    });
  } catch (e) {
    const error = new Error("000 No connection");
    error.status = 0;
    console.log(e);
    throw error;
  }

  if (res.status < 200 || res.status > 299) {
    const error = new Error(await res.text());
    error.status = res.status;

    if (res.status === 401) {
      atExpired &&
        Vue.prototype.$showError({ message: i18n.t(atExpired) }, false);
      setAt && logout();
    }

    throw error;
  }
  if (res.status === 200) {
    return res.json();
  } else {
    throw new Error(res.status);
  }
}

/* api */
export function saveToken(at) {
  store.commit("bd/setAt", at);
  sessionStorage.setItem("bdAt", at);
}

export function getToken() {
  const { at } = store.state.bd;
  if (at) return at;
  const bdAt = sessionStorage.getItem("bdAt");
  if (bdAt && bdAt !== "null") {
    store.commit("bd/setAt", bdAt);
    return bdAt;
  }
}

export async function login(code) {
  const { access_token } = await fetchUtil(
    "/api/bd/login",
    {
      method: "POST",
      body: { code },
    },
    false,
    "baiduNetdisk.bindFail"
  );
  if (access_token) {
    saveToken(access_token);
    //切换用户
    store.commit("bd/setChange", true);
  }
}

export function logout() {
  store.commit("bd/setAt", "");
  store.commit("bd/setUser", "");
  store.commit("bd/updateReq", {});
  sessionStorage.setItem("bdAt", null);

  if (router.currentRoute.path !== "/baidu-netdisk/") {
    router.push({ path: "/baidu-netdisk" });
  }
}

// Format the chunk size in bytes to user-friendly format
function formatBytes(bytes = 0) {
  const units = ["B", "KB", "MB", "GB"];
  let size = bytes;
  let unitIndex = 0;
  while (size >= 1024 && unitIndex < units.length - 1) {
    size /= 1024;
    unitIndex++;
  }
  return `${size.toFixed(1)}${units[unitIndex]}`;
}

export async function fetchUserInfo() {
  // 1. 获取用户信息 2. 验证 bd assesstoken 是否有误
  const {
    baidu_name: name,
    has_used,
    is_vip = 0,
    total_cap,
  } = await fetchUtil("/api/bd/user-info", {
    method: "POST",
  });
  const vipMap = {
    0: "baiduNetdisk.ordinaryUser",
    1: "baiduNetdisk.ordinaryMember",
    2: "baiduNetdisk.superMember",
  };
  const data = {
    name,
    type: vipMap[is_vip],
    hasUsed: formatBytes(has_used),
    totolCap: formatBytes(total_cap),
    usedPercent: Math.round((has_used / total_cap) * 100),
  };
  store.commit("bd/setUser", data);
  store.commit("setReload", true);
}

export async function fetchDir(url = "/") {
  let path = removePrefix(url);
  if (path.slice(-1) !== "/") path += "/";
  let rooturl = `/baidu-netdisk${path}`;
  const { errno, list } = await fetchUtil("/api/bd/dir", {
    method: "POST",
    body: { path },
  });

  if (errno === -6) {
    throw { status: 401 };
  } else if (errno === -7) {
    throw { status: 403 };
  } else if (errno === -9) {
    throw { status: 404 };
  }

  if (!list) return;

  let nameArr = path.split("/"),
    name = nameArr[nameArr.length - 2] || i18n.t("sidebar.baiduNetdisk");

  let numDirs = 0,
    numFiles = 0,
    mtime = 0;

  const items = list.map((item, index) => {
    let {
      fs_id: fsId,
      real_category: extension,
      isdir,
      server_mtime,
      server_filename: name,
      path,
      size,
    } = item;
    let isDir = isdir === 1,
      url = "",
      modified = new Date(server_mtime * 1000).toISOString(),
      type = undefined;

    if (isDir) {
      numDirs++;
      url = `${rooturl}${name}/`;
      type = "";
    } else {
      numFiles++;
    }
    if (mtime < server_mtime) mtime = server_mtime;

    return {
      fsId,
      extension,
      index,
      isDir,
      isSymlink: undefined,
      mode: undefined,
      modified,
      name,
      path,
      size,
      type,
      url,
    };
  });

  const res = {
    extension: "",
    isDir: true,
    isSymlink: undefined,
    items,
    mode: undefined,
    modified: new Date(mtime * 1000).toISOString(),
    name,
    numDirs,
    numFiles,
    path,
    size: undefined,
    sorting: {},
    type: "",
    url: rooturl,
  };

  store.commit("bd/updateReq", res);
}

// path、is_dir、fs_id、target_path
export function fetchDownload(data) {
  return fetchUtil("/api/bd/download", {
    method: "POST",
    body: data,
  });
}

export function fetchProgress() {
  return fetchUtil("/api/bd/progress", {
    method: "POST",
  });
}

export function deleteProgress(data) {
  return fetchUtil("/api/bd/progress", {
    method: "DELETE",
    body: data,
  });
}

// file_name: string
export function stopProgress(data) {
  return fetchUtil("/api/bd/progress/stop", {
    method: "PATCH",
    body: data,
  });
}

export function continueProgress(data) {
  return fetchUtil("/api/bd/progress/restart", {
    method: "PATCH",
    body: data,
  });
}

export function cancelProgress(data) {
  return fetchUtil("/api/bd/progress/cancel", {
    method: "PATCH",
    body: data,
  });
}

export async function getAccessToken() {
  if (sessionStorage.getItem("bdAt")) {
    return;
  }
  const res = await fetchUtil(
    "/api/bd/access-token",
    {
      method: "GET",
    },
    false
  );
  const { access_token } = res.data;
  if (access_token) saveToken(access_token);
}
