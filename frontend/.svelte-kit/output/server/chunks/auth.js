import { w as writable } from "./index.js";
function createAuthStore() {
  const { subscribe, set, update } = writable({
    authenticated: false,
    hasCredentials: false,
    loading: true,
    error: null
  });
  return {
    subscribe,
    async checkStatus() {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        const response = await fetch("/api/auth/status", {
          credentials: "include"
        });
        if (response.ok) {
          const data = await response.json();
          set({
            authenticated: data.authenticated,
            hasCredentials: data.hasCredentials,
            loading: false,
            error: null
          });
          return data;
        } else {
          throw new Error("Failed to check auth status");
        }
      } catch (error) {
        update((state) => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : "Unknown error"
        }));
        return null;
      }
    },
    async registerPasskey() {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        const beginRes = await fetch("/api/auth/register/begin", {
          method: "POST",
          credentials: "include"
        });
        if (!beginRes.ok) {
          throw new Error("Failed to begin registration");
        }
        const options = await beginRes.json();
        options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
        options.publicKey.user.id = base64urlToBuffer(options.publicKey.user.id);
        if (options.publicKey.excludeCredentials) {
          options.publicKey.excludeCredentials = options.publicKey.excludeCredentials.map(
            (cred) => ({
              ...cred,
              id: base64urlToBuffer(cred.id)
            })
          );
        }
        const credential = await navigator.credentials.create({
          publicKey: options.publicKey
        });
        if (!credential) {
          throw new Error("Failed to create credential");
        }
        const attestationResponse = credential.response;
        const credentialData = {
          id: credential.id,
          rawId: bufferToBase64url(credential.rawId),
          type: credential.type,
          response: {
            clientDataJSON: bufferToBase64url(attestationResponse.clientDataJSON),
            attestationObject: bufferToBase64url(attestationResponse.attestationObject)
          }
        };
        const finishRes = await fetch("/api/auth/register/finish", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(credentialData),
          credentials: "include"
        });
        if (!finishRes.ok) {
          const errorText = await finishRes.text();
          throw new Error(errorText || "Failed to complete registration");
        }
        set({
          authenticated: true,
          hasCredentials: true,
          loading: false,
          error: null
        });
        return true;
      } catch (error) {
        update((state) => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : "Registration failed"
        }));
        return false;
      }
    },
    async login() {
      update((state) => ({ ...state, loading: true, error: null }));
      try {
        const beginRes = await fetch("/api/auth/login/begin", {
          method: "POST",
          credentials: "include"
        });
        if (!beginRes.ok) {
          throw new Error("Failed to begin login");
        }
        const options = await beginRes.json();
        options.publicKey.challenge = base64urlToBuffer(options.publicKey.challenge);
        if (options.publicKey.allowCredentials) {
          options.publicKey.allowCredentials = options.publicKey.allowCredentials.map(
            (cred) => ({
              ...cred,
              id: base64urlToBuffer(cred.id)
            })
          );
        }
        const assertion = await navigator.credentials.get({
          publicKey: options.publicKey
        });
        if (!assertion) {
          throw new Error("Failed to get credential");
        }
        const assertionResponse = assertion.response;
        const assertionData = {
          id: assertion.id,
          rawId: bufferToBase64url(assertion.rawId),
          type: assertion.type,
          response: {
            clientDataJSON: bufferToBase64url(assertionResponse.clientDataJSON),
            authenticatorData: bufferToBase64url(assertionResponse.authenticatorData),
            signature: bufferToBase64url(assertionResponse.signature),
            userHandle: assertionResponse.userHandle ? bufferToBase64url(assertionResponse.userHandle) : null
          }
        };
        const finishRes = await fetch("/api/auth/login/finish", {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify(assertionData),
          credentials: "include"
        });
        if (!finishRes.ok) {
          const errorText = await finishRes.text();
          throw new Error(errorText || "Failed to complete login");
        }
        set({
          authenticated: true,
          hasCredentials: true,
          loading: false,
          error: null
        });
        return true;
      } catch (error) {
        update((state) => ({
          ...state,
          loading: false,
          error: error instanceof Error ? error.message : "Login failed"
        }));
        return false;
      }
    },
    async logout() {
      try {
        await fetch("/api/auth/logout", {
          method: "POST",
          credentials: "include"
        });
      } catch {
      }
      set({
        authenticated: false,
        hasCredentials: true,
        loading: false,
        error: null
      });
    },
    clearError() {
      update((state) => ({ ...state, error: null }));
    }
  };
}
function base64urlToBuffer(base64url) {
  const base64 = base64url.replace(/-/g, "+").replace(/_/g, "/");
  const padded = base64.padEnd(base64.length + (4 - base64.length % 4) % 4, "=");
  const binary = atob(padded);
  const buffer = new ArrayBuffer(binary.length);
  const bytes = new Uint8Array(buffer);
  for (let i = 0; i < binary.length; i++) {
    bytes[i] = binary.charCodeAt(i);
  }
  return buffer;
}
function bufferToBase64url(buffer) {
  const bytes = new Uint8Array(buffer);
  let binary = "";
  for (let i = 0; i < bytes.length; i++) {
    binary += String.fromCharCode(bytes[i]);
  }
  return btoa(binary).replace(/\+/g, "-").replace(/\//g, "_").replace(/=/g, "");
}
createAuthStore();
