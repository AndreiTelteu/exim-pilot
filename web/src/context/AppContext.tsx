import { createContext, useContext, useReducer, ReactNode } from "react";

// State interface
interface AppState {
  isLoading: boolean;
  error: string | null;
  notifications: Notification[];
  user: User | null;
  connectionStatus: "connected" | "disconnected" | "connecting";
}

interface User {
  id: string;
  username: string;
  role: string;
}

interface Notification {
  id: string;
  type: "success" | "error" | "warning" | "info";
  message: string;
  timestamp: number;
}

// Action types
type AppAction =
  | { type: "SET_LOADING"; payload: boolean }
  | { type: "SET_ERROR"; payload: string | null }
  | {
      type: "ADD_NOTIFICATION";
      payload: Omit<Notification, "id" | "timestamp">;
    }
  | { type: "REMOVE_NOTIFICATION"; payload: string }
  | { type: "SET_USER"; payload: User | null }
  | {
      type: "SET_CONNECTION_STATUS";
      payload: "connected" | "disconnected" | "connecting";
    };

// Initial state
const initialState: AppState = {
  isLoading: false,
  error: null,
  notifications: [],
  user: null,
  connectionStatus: "disconnected",
};

// Reducer
function appReducer(state: AppState, action: AppAction): AppState {
  switch (action.type) {
    case "SET_LOADING":
      return { ...state, isLoading: action.payload };

    case "SET_ERROR":
      return { ...state, error: action.payload };

    case "ADD_NOTIFICATION":
      const newNotification: Notification = {
        ...action.payload,
        id: Date.now().toString(),
        timestamp: Date.now(),
      };
      return {
        ...state,
        notifications: [...state.notifications, newNotification],
      };

    case "REMOVE_NOTIFICATION":
      return {
        ...state,
        notifications: state.notifications.filter(
          (n) => n.id !== action.payload
        ),
      };

    case "SET_USER":
      return { ...state, user: action.payload };

    case "SET_CONNECTION_STATUS":
      return { ...state, connectionStatus: action.payload };

    default:
      return state;
  }
}

// Context
interface AppContextType {
  state: AppState;
  dispatch: React.Dispatch<AppAction>;
  actions: {
    setLoading: (loading: boolean) => void;
    setError: (error: string | null) => void;
    addNotification: (
      notification: Omit<Notification, "id" | "timestamp">
    ) => void;
    removeNotification: (id: string) => void;
    setUser: (user: User | null) => void;
    setConnectionStatus: (
      status: "connected" | "disconnected" | "connecting"
    ) => void;
  };
}

const AppContext = createContext<AppContextType | undefined>(undefined);

// Provider component
interface AppProviderProps {
  children: ReactNode;
}

export function AppProvider({ children }: AppProviderProps) {
  const [state, dispatch] = useReducer(appReducer, initialState);

  const actions = {
    setLoading: (loading: boolean) =>
      dispatch({ type: "SET_LOADING", payload: loading }),
    setError: (error: string | null) =>
      dispatch({ type: "SET_ERROR", payload: error }),
    addNotification: (notification: Omit<Notification, "id" | "timestamp">) =>
      dispatch({ type: "ADD_NOTIFICATION", payload: notification }),
    removeNotification: (id: string) =>
      dispatch({ type: "REMOVE_NOTIFICATION", payload: id }),
    setUser: (user: User | null) =>
      dispatch({ type: "SET_USER", payload: user }),
    setConnectionStatus: (
      status: "connected" | "disconnected" | "connecting"
    ) => dispatch({ type: "SET_CONNECTION_STATUS", payload: status }),
  };

  return (
    <AppContext.Provider value={{ state, dispatch, actions }}>
      {children}
    </AppContext.Provider>
  );
}

// Hook to use the context
export function useApp() {
  const context = useContext(AppContext);
  if (context === undefined) {
    throw new Error("useApp must be used within an AppProvider");
  }
  return context;
}
