import { useEffect } from "react";
import { startRegistration } from "@simplewebauthn/browser";

export default function Register() {
    useEffect(() => {
        async function register() {
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
              body: JSON.stringify(registerResponse),
            });
    
            // Wait for the results of verification
            const verificationJSON = await verificationResp.json();
            console.log(verificationJSON);
          } catch (error) {
            console.error(error);
          }
        }
        register();
      }, []);

    return (
        <div>
            <h1>Register</h1>
        </div>
    );
}