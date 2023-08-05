import { Component } from '@angular/core';
import { FormControl, FormGroup } from '@angular/forms';
import { NewSchedule } from 'src/app/models/schedule.model';
import { ScheduleService } from 'src/app/services/schedule.service';
import { SchedulesService } from 'src/app/services/schedules.service';

@Component({
  selector: 'app-new-schedule',
  templateUrl: './new-schedule.component.html',
  styleUrls: ['./new-schedule.component.css']
})
export class NewScheduleComponent {
  schedule = new FormGroup({
    cron: new FormControl(''),
    hook: new FormControl(''),
    data: new FormControl(''),
    till: new FormControl('2023-08-08T16:37:50Z'),
    active: new FormControl(true),
  });

  hooks: string[] = [];

  constructor(private schedulesService: SchedulesService, private scheduleService: ScheduleService) { }

  ngOnInit(): void {
    this.loadHooks();
  }

  loadHooks() {
    this.schedulesService.getHooks().subscribe(
      data => {
        this.hooks = data;
      }
    );
  }

  createSchedule() {
    console.log(this.schedule.value);
    this.scheduleService.createSchedule(this.schedule.value).subscribe(
      data => {

      }
    )
  }
}
