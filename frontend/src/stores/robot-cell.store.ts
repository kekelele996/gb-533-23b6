import { inject, Injectable, signal } from '@angular/core';
import { finalize } from 'rxjs';
import { RobotCellApi } from '../api/robot-cell';
import { apiErrorIssues, apiErrorMessage } from '../utils/api-error';
import { CreateRobotCell, RobotCell, UpdateRobotCell } from '../types/robot-cell';

@Injectable({ providedIn: 'root' })
export class RobotCellStore {
  private readonly api = inject(RobotCellApi);
  readonly items = signal<RobotCell[]>([]);
  readonly selected = signal<RobotCell | null>(null);
  readonly loading = signal(false);
  readonly error = signal('');
  readonly freezeIssues = signal<string[]>([]);

  load(): void {
    this.loading.set(true);
    this.error.set('');
    this.freezeIssues.set([]);
    this.api.list().pipe(finalize(() => this.loading.set(false))).subscribe({
      next: ({ data }) => {
        this.items.set(data);
        if (!this.selected() && data.length) this.selected.set(data[0]);
      },
      error: (error) => this.error.set(apiErrorMessage(error)),
    });
  }

  choose(cell: RobotCell): void { this.selected.set(cell); this.freezeIssues.set([]); }

  create(payload: CreateRobotCell, done?: () => void): void {
    this.mutate(this.api.create(payload), done);
  }

  update(id: number, payload: UpdateRobotCell, done?: () => void): void {
    this.mutate(this.api.update(id, payload), done);
  }

  freeze(cell: RobotCell): void {
    this.mutate(this.api.freeze(cell.id), undefined, (error) => {
      this.freezeIssues.set(apiErrorIssues(error));
      this.error.set(apiErrorMessage(error));
    });
  }
  deactivate(cell: RobotCell): void { this.mutate(this.api.deactivate(cell.id)); }

  private mutate(request: ReturnType<RobotCellApi['create']>, done?: () => void, onError?: (error: unknown) => void): void {
    this.loading.set(true);
    this.error.set('');
    this.freezeIssues.set([]);
    request.pipe(finalize(() => this.loading.set(false))).subscribe({
      next: ({ data }) => {
        this.items.update((items) => [data, ...items.filter((item) => item.id !== data.id)].sort((a, b) => a.cell_code.localeCompare(b.cell_code)));
        this.selected.set(data);
        done?.();
      },
      error: (error) => {
        if (onError) onError(error);
        else this.error.set(apiErrorMessage(error));
      },
    });
  }
}
