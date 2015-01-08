import ENV from 'admin/config/environment';
import DS from 'ember-data';

var ApplicationAdapter = DS.RESTAdapter;

if ("DS" in ENV) {
  ApplicationAdapter = DS.RESTAdapter.extend(ENV.DS);
}

export default ApplicationAdapter;