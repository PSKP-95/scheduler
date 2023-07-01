import { ScheduleStatus } from "../enums/schedule-status";

export class Schedule {
  id: string = '00000000-0000-0000-0000-000000000000';
  cron: string = '* * * * *';
  hook: string = "";
  active: boolean = true;
  till!: Date;
  data: string = '';
  created_at!: Date;
  last_modified!: Date;
}

export class Page {
  number: number = 0;
  size: number = 10;
  totalPages: number = 0;
  totalElements: number = 0;
}

export class History {
  occurence_id: number = 0;
  schedule: string = '00000000-0000-0000-0000-000000000000';
  status: ScheduleStatus = ScheduleStatus.PENDING;
  details: string = '';
  manual: boolean = false;
  scheduled_at!: Date;
  started_at!: Date;
  completed_at!: Date;
}

export class SchedulesResponse {
  schedules: Schedule[];
  page: Page;
}

export class ScheduleHistoryResponse {
  history: History[];
  page: Page;
}
