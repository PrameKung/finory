type ApiErrorBody = {
  code?: unknown;
  error?: unknown;
  message?: unknown;
};

export type ApiErrorDetails = {
  code?: string;
  message?: string;
};

type ApiErrorMessageOptions = {
  defaultMessage: string;
  codeMessages?: Readonly<Record<string, string>>;
  statusMessages?: Readonly<Record<number, string>>;
};

export class ApiError extends Error {
  readonly status: number;
  readonly code?: string;

  constructor(
    status: number,
    code?: string,
    options: { cause?: unknown; message?: string } = {},
  ) {
    super(options.message ?? code ?? `API request failed with status ${status}`, {
      cause: options.cause,
    });
    this.name = "ApiError";
    this.status = status;
    this.code = code;
  }
}

function stringValue(value: unknown) {
  return typeof value === "string" && value.trim()
    ? value.trim()
    : undefined;
}

export function parseApiErrorBody(body: unknown): ApiErrorDetails {
  if (!body || typeof body !== "object") {
    return {};
  }

  const { code, error, message } = body as ApiErrorBody;

  if (error && typeof error === "object") {
    const nestedError = error as ApiErrorBody;

    return {
      code: stringValue(nestedError.code),
      message: stringValue(nestedError.message),
    };
  }

  return {
    code: stringValue(error) ?? stringValue(code),
    message: stringValue(message),
  };
}

export function getApiErrorMessage(
  error: unknown,
  {
    defaultMessage,
    codeMessages = {},
    statusMessages = {},
  }: ApiErrorMessageOptions,
) {
  if (!(error instanceof ApiError)) {
    return defaultMessage;
  }

  if (error.code && codeMessages[error.code]) {
    return codeMessages[error.code];
  }

  if (statusMessages[error.status]) {
    return statusMessages[error.status];
  }

  if (error.code === "network_error") {
    return "We could not reach the server. Check your connection and try again.";
  }

  if (error.status === 401) {
    return "Your session has expired. Sign in and try again.";
  }

  if (error.status === 403) {
    return "You do not have permission to do that.";
  }

  if (error.status === 429) {
    return "Too many requests. Wait a moment and try again.";
  }

  if (error.status >= 500) {
    return "The service is temporarily unavailable. Please try again.";
  }

  return defaultMessage;
}
