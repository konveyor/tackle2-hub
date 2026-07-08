import React, { useState } from "react";
import {
  LoginForm,
  LoginPage as PFLoginPage,
  ListItem,
  ListVariant,
} from "@patternfly/react-core";
import { ExclamationCircleIcon } from "@patternfly/react-icons";

import { brandingStrings } from "./branding";
import type { LoginConfig } from "./types";

interface UserLoginPageProps {
  config: LoginConfig;
}

// UserLoginPage renders the username/password login form using PatternFly 6's
// LoginPage component. The form action and any federated IdP button are
// provided via the hub-injected config; branding comes from the build-time bundle.
export const UserLoginPage: React.FC<UserLoginPageProps> = ({ config }) => {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");

  const { formAction, errorMessage, federatedIdp } = config;
  const { loginPage, styles, application } = brandingStrings;

  // TODO Add support for other system error messages (e.g. authRequestId timeout)
  const hasError = Boolean(errorMessage);
  const helperText = hasError ? errorMessage : undefined;

  // TODO Add support for >1 federated IdP button
  // Federated IdP link rendered as a footer list item when configured.
  const socialLoginContent = federatedIdp ? (
    <ListItem>
      <a href={federatedIdp.loginUrl}>Sign in with {federatedIdp.name}</a>
    </ListItem>
  ) : undefined;

  return (
    <PFLoginPage
      loginTitle={loginPage?.title ?? "Log in to your account"}
      loginSubtitle={loginPage?.subtitle}
      backgroundImgSrc={styles?.backgroundImage ?? ""}
      brandImgProps={{
        src: styles?.brandImage ?? "",
        alt: application?.name ?? application?.title ?? "",
        widths: { default: "260px" }
      }}
      textContent={loginPage?.textContent ?? ""}
      socialMediaLoginContent={socialLoginContent}
      socialMediaLoginAriaLabel={
        federatedIdp ? "Alternative login options" : undefined
      }
      footerListVariants={ListVariant.inline}
    >
      <LoginForm
        method="post"
        action={formAction ?? ""}
        id="pf-login-form-id"
        usernameLabel="Username"
        usernameValue={username}
        onChangeUsername={(_e, v) => setUsername(v)}
        isValidUsername={!hasError}
        passwordLabel="Password"
        passwordValue={password}
        onChangePassword={(_e, v) => setPassword(v)}
        isValidPassword={!hasError}
        showHelperText={hasError}
        helperText={helperText}
        helperTextIcon={<ExclamationCircleIcon />}
        loginButtonLabel="Log in"
        isShowPasswordEnabled
        // onLoginButtonClick is omitted; native form submit handles the POST
      />
    </PFLoginPage>
  );
};
