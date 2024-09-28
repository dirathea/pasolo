import { useEffect } from "react";
import { startAuthentication } from "@simplewebauthn/browser";

export default function Login() {

    useEffect(() => {
        async function login() {
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
    
            // Wait for the results of verification
            const verificationJSON = await verificationResp.json();
            console.log(verificationJSON);

            const urlParams = new URLSearchParams(window.location.search);
            const redirectUrl = urlParams.get('redirect') || '/';
            window.location.href = redirectUrl;

          } catch (error) {
            window.location.href = '/register';
          }
        }
        login();
      }, []);

    return (
        <div>
            <h1>Login</h1>
        </div>
    );
}