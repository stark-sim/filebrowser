import Vue from "vue";
import store from "@/store";
import router from "@/router";
import { baseURL } from "@/utils/constants";
import { removePrefix, fetchURL, fetchJSON } from "@/api/utils";
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
  opts.body = opts.body || {};
  let { headers, body, ...rest } = opts;

  body = { ...body };

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
    // const error = new Error("000 No connection");
    // error.status = 0;

    throw e;
  }

  if (res.status < 200 || res.status > 299) {
    const error = new Error(await res.text());
    error.status = res.status;

    if (res.status === 401) {
      atExpired && Vue.prototype.$showError(i18n.t(atExpired), false, 1500);
      setAt && logout();
    }

    throw error;
  }
  if (res.status >= 200 && res.status <= 299) {
    return await res;
  } else {
    throw new Error(res.status);
  }
}

export async function fetchDir(url = "/") {
  let path = removePrefix(url);
  if (path.slice(-1) !== "/") path += "/";
  let rooturl = `/cephalon-cloud${path}`;
  const res1 = await fetchJSON(`/api/cd/dir`, {
    method: "GET",
  });
  const { errno, list } = res1.data;
  if (errno === -6) {
    throw { status: 401 };
  } else if (errno === -7) {
    throw { status: 403 };
  } else if (errno === -9) {
    throw { status: 404 };
  }

  if (!list) return;
  let nameArr = path.split("/"),
    name = nameArr[nameArr.length - 2] || i18n.t("sidebar.cephalonCloud");

  let numDirs = 0,
    numFiles = 0;
  // mtime = 0;
  const items = list.map((item, index) => {
    let { create_at, deleted_at, md5, size, updated_at, name } = item;
    let isDir = false;
    let url = "";
    //   modified = new Date(server_mtime * 1000).toISOString(),

    numFiles++;
    // }
    // if (mtime < server_mtime) mtime = server_mtime;

    return {
      // fsId,
      // extension,
      index,
      isDir,
      isSymlink: undefined,
      mode: undefined,
      // modified,
      name,
      // path,
      size,
      type: "text",
      url,
      create_at,
      deleted_at,
      md5,
      updated_at,
    };
  });

  const res = {
    extension: "",
    isDir: true,
    isSymlink: undefined,
    items,
    mode: undefined,
    // modified: new Date(mtime * 1000).toISOString(),
    name,
    numDirs,
    numFiles,
    path,
    size: undefined,
    sorting: {},
    type: "",
    url: rooturl,
  };
  store.commit("cep/updateReq", res);
}

// path、is_dir、fs_id、target_path
export function fetchDownload(data) {
  return fetchUtil("/api/cd/download", {
    method: "POST",
    body: data,
  });
}

// export function fetchProgress() {
//   return fetchUtil("/api/bd/progress", {
//     method: "POST",
//   });
// }

export async function usage() {
  const res = await fetchURL(`/api/cd/user-space`, {
    method: "GET",
  });
  return await res.json();
}
