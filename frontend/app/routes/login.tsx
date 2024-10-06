import { useEffect } from "react";
import { startAuthentication } from "@simplewebauthn/browser";

export default function Login() {
  async function login() {
    const urlParams = new URLSearchParams(window.location.search);
    const redirectUrl = urlParams.get("rd") || "/";

    const register = await fetch("/auth/login", {
      method: "GET",
    });
    try {
      const response = await register.json();
      console.log(response);
      const registerResponse = await startAuthentication(response);
      console.log(registerResponse);

      // POST the response to the endpoint that calls
      // @simplewebauthn/server -> verifyRegistrationResponse()
      const verificationResp = await fetch("/auth/login", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(registerResponse),
      });

      if (!verificationResp.ok) {
        throw new Error("Failed to verify login");
      } else {
        window.location.href = redirectUrl;
      }
    } catch (error) {
      console.error(error);
      window.location.href = "/register?rd=" + redirectUrl;
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-inherit">
      <div className="p-8 rounded shadow-md w-full max-w-md flex justify-center">
      <form
        onSubmit={async (e) => {
        e.preventDefault();
        await login();
        }}
      >
        <div className="flex justify-center">
        <button
          type="submit"
          className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
        >
          Login with passkey
        </button>
        </div>
      </form>
      </div>
    </div>
  );
}
