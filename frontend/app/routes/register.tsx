import { useEffect } from "react";
import { startRegistration } from "@simplewebauthn/browser";

export default function Register() {

  async function register(password: string) {

    const urlParams = new URLSearchParams(window.location.search);
    const redirectUrl = urlParams.get('rd') || '/';

    const register = await fetch("/auth/register", {
      method: "GET",
    });
    try {
      const response = await register.json();
      console.log(response);
      const registerResponse = await startRegistration(response);
      console.log(registerResponse);

      // POST the response to the endpoint that calls
      // @simplewebauthn/server -> verifyRegistrationResponse()
      const verificationResp = await fetch("/auth/register", {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify({
          credential: registerResponse,
          password,
        }),
      });

      window.location.href = '/login?rd=' + redirectUrl;
    } catch (error) {
      console.error(error);
    }
  }

    return (
        <div className="flex items-center justify-center min-h-screen bg-inherit">
            <div className="p-8 rounded shadow-md w-full max-w-md">
          <h1 className="text-2xl font-bold mb-6 text-center">Register</h1>
          <form
            onSubmit={async (e) => {
              e.preventDefault();
              const password = (e.target as HTMLFormElement).elements.namedItem("password") as HTMLInputElement;
              await register(password.value);
            }}
          >
            <div className="mb-4">
              <label className="block text-gray-700 text-sm font-bold mb-2">
                Password:
                <input
            type="password"
            name="password"
            required
            className="shadow appearance-none border rounded w-full py-2 px-3 text-gray-700 leading-tight focus:outline-none focus:shadow-outline"
                />
              </label>
            </div>
            <div className="flex items-center justify-between">
              <button
                type="submit"
                className="bg-blue-500 hover:bg-blue-700 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline"
              >
                Register
              </button>
            </div>
          </form>
            </div>
        </div>
    );
}