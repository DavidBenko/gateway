import Ember from "ember";

var IndexRoute = Ember.Route.extend({
  beforeModel: function() {
    this.transitionTo('admin');
  }
});

export default IndexRoute;