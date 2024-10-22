import { startAuthentication } from "@simplewebauthn/browser";
import { useNavigate } from "@remix-run/react";
import { useState } from "react";

export default function Login() {
  const navigate = useNavigate();
  const [isLoggedIn, setIsLoggedIn] = useState(false);
  const urlParams = new URLSearchParams(window.location.search);
  const redirectUrl = urlParams.get("rd") || "/";

  async function login() {
    const register = await fetch("/auth/login", {
      method: "GET",
    });
    try {
      const response = await register.json();
      const registerResponse = await startAuthentication(response);

      // POST the response to the endpoint that calls
      // @simplewebauthn/server -> verifyRegistrationResponse()
      const queryParams = new URLSearchParams({
        rd: redirectUrl,
      });
      const verificationUrl = `/auth/login?${queryParams.toString()}`;
      const verificationResp = await fetch(verificationUrl, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(registerResponse),
      });

      if (!verificationResp.ok) {
        throw new Error("Failed to verify login");
      }

      setIsLoggedIn(verificationResp.ok);
    } catch (error) {
      console.error(error);
      navigate("/register?rd=" + redirectUrl);
    }
  }

  return (
    <div className="flex items-center justify-center min-h-screen bg-inherit">
      <div className="p-8 rounded shadow-md w-full max-w-md flex justify-center">
        {isLoggedIn ? (
          <a
            href={redirectUrl}
            className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
          >
            Continue
          </a>
        ) : (
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
        )}
      </div>
    </div>
  );
}
