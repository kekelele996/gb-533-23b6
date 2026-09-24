import { HttpErrorResponse } from '@angular/common/http';
import { ApiFailure } from '../types/api';
import { FreezeCheckIssue } from '../types/robot-cell';

export function apiFailure(error: unknown): ApiFailure | undefined {
  if (!(error instanceof HttpErrorResponse)) return undefined;
  return error.error as ApiFailure | undefined;
}

export function apiErrorMessage(error: unknown): string {
  if (!(error instanceof HttpErrorResponse)) return 'Unexpected application error';
  const failure = error.error as ApiFailure | undefined;
  const message = failure?.error?.message ?? `Request failed with HTTP ${error.status}`;
  const requestId = failure?.request_id ? ` · Request ${failure.request_id}` : '';
  return message + requestId;
}

export function apiErrorCode(error: unknown): string | undefined {
  return apiFailure(error)?.error?.code;
}

// Extracts the per-issue list returned with a 409 freeze_checks_failed response.
export function apiFreezeIssues(error: unknown): FreezeCheckIssue[] {
  const details = apiFailure(error)?.error?.details;
  if (!details || typeof details !== 'object') return [];
  const issues = (details as { issues?: unknown }).issues;
  if (!Array.isArray(issues)) return [];
  return issues.filter((item): item is FreezeCheckIssue =>
    !!item && typeof item === 'object' && typeof (item as FreezeCheckIssue).kind === 'string'
      && typeof (item as FreezeCheckIssue).message === 'string');
}
